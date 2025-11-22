package higher

import (
	"encoding/json"
	"errors"
	"net"
	"sync"
	"time"
)

const (
	MessageTypeBroadcast           = "server_broadcast"
	MessageTypePublish             = "server_publish"
	MessageTypeDirectMessage       = "server_direct_message"
	ClientMessageTypeBroadcast     = "client_broadcast"
	ClientMessageTypePublish       = "client_publish"
	ClientMessageTypeDirectMessage = "client_direct_message"
	ClientMessageTypeGetPresence   = "client_get_presence"
)

type HigherSockd struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func NewSockd() *HigherSockd {
	return &HigherSockd{
		rooms: make(map[string]*Room),
		mu:    sync.RWMutex{},
	}
}

func newRoom(name string) *Room {
	r := &Room{
		name:       name,
		disconnect: make(chan int64, 32),
		broadcast:  make(chan []byte, 32),
		publish:    make(chan publishEvent, 32),
		directMsg:  make(chan directMessageEvent, 32),
		topics:     make(map[string]map[int64]bool),
		tLock:      sync.RWMutex{},
		sessions:   make(map[int64]*session),
		sLock:      sync.RWMutex{},
	}

	go r.run()

	return r
}

func (s *HigherSockd) AddConn(userId int64, conn net.Conn, connId int64, roomName string) (int64, error) {
	s.mu.Lock()
	room, exists := s.rooms[roomName]
	if !exists {
		room = newRoom(roomName)
		s.rooms[roomName] = room
	}
	s.mu.Unlock()

	sess := &session{
		room:   room,
		connId: connId,
		userId: userId,
		conn:   conn,
		send:   make(chan []byte, 16),
	}

	room.sLock.Lock()
	existingSess := room.sessions[sess.connId]
	room.sessions[sess.connId] = sess
	room.sLock.Unlock()

	if existingSess != nil {
		existingSess.teardown()
	}

	go sess.writePump()
	go sess.readPump()

	return sess.connId, nil
}

func (s *HigherSockd) RemoveConn(userId int64, connId int64, roomName string) error {
	s.mu.RLock()
	room, exists := s.rooms[roomName]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	tcan := time.After(time.Second * 10)

	select {
	case room.disconnect <- connId:
		return nil
	case <-tcan:
		return errors.New("room is very busy or dead")
	}
}

// Broadcast sends a message to all users in the room
func (s *HigherSockd) Broadcast(roomName string, message []byte) error {
	s.mu.RLock()
	room, exists := s.rooms[roomName]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	msg := Message{
		Id:   time.Now().UnixNano(),
		Type: MessageTypeBroadcast,
		Data: message,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	room.handleBroadcast(data, time.Second*2)

	return nil
}

// Publish sends a message to all subscribers of a topic in the room
func (s *HigherSockd) Publish(roomName string, topicName string, message []byte) error {
	s.mu.RLock()
	room, exists := s.rooms[roomName]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	msg := Message{
		Id:   time.Now().UnixNano(),
		Type: "server_publish",
		Data: message,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	room.handlePublish(topicName, data, time.Second*5)

	return nil

}

// DirectMessage sends a message to a specific user
func (s *HigherSockd) DirectMessage(roomName string, targetConnId int64, message []byte) error {
	s.mu.RLock()
	room, exists := s.rooms[roomName]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	msg := Message{
		Id:   time.Now().UnixNano(),
		Type: "server_direct_message",
		Data: message,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	room.handleDirectMessage(targetConnId, data, time.Second*5)

	return nil

}

// Subscribe adds a connection to a topic subscription
func (s *HigherSockd) Subscribe(roomName string, topicName string, connId int64) error {
	s.mu.RLock()
	room, exists := s.rooms[roomName]
	s.mu.RUnlock()

	if !exists {
		return errors.New("room not found")
	}

	room.tLock.Lock()
	if room.topics[topicName] == nil {
		room.topics[topicName] = make(map[int64]bool)
	}
	room.topics[topicName][connId] = true
	room.tLock.Unlock()

	return nil
}

// Unsubscribe removes a connection from a topic subscription
func (s *HigherSockd) Unsubscribe(roomName string, topicName string, connId int64) error {
	s.mu.RLock()
	room, exists := s.rooms[roomName]
	s.mu.RUnlock()

	if !exists {
		return errors.New("room not found")
	}

	room.tLock.Lock()
	if subMap, ok := room.topics[topicName]; ok {
		delete(subMap, connId)
		if len(subMap) == 0 {
			delete(room.topics, topicName)
		}
	}
	room.tLock.Unlock()

	return nil
}
