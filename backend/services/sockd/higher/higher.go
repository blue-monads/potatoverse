package higher

import (
	"errors"
	"net"
	"sync"
)

const (
	MessageTypeBroadcast           = "sbroadcast"
	MessageTypePublish             = "spublish"
	MessageTypeDirectMessage       = "sdirect"
	ClientMessageTypeBroadcast     = "cbroadcast"
	ClientMessageTypePublish       = "cpublish"
	ClientMessageTypeDirectMessage = "cdirect"
	ClientMessageTypeGetPresence   = "cpresence"
	ClientMessageTypeCommand       = "ccommand"
)

type HigherSockd struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func New() HigherSockd {
	return HigherSockd{
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

func (h *HigherSockd) getRoom(roomName string, createIfNotExists bool) *Room {
	h.mu.RLock()
	room, exists := h.rooms[roomName]
	h.mu.RUnlock()
	if !exists {
		if createIfNotExists {
			room = newRoom(roomName)
			h.mu.Lock()
			defer h.mu.Unlock()
			sneakyRoom := h.rooms[roomName]
			if sneakyRoom != nil {
				return sneakyRoom
			}
			h.rooms[roomName] = room

			go room.run()
		}

		return room
	}
	return room
}

func (s *HigherSockd) AddConn(userId int64, conn net.Conn, connId int64, roomName string) (int64, error) {
	room := s.getRoom(roomName, true)
	if room == nil {
		return 0, errors.New("room not found")
	}

	return room.AddConn(userId, conn, connId)

}

// Broadcast sends a message to all users in the room
func (s *HigherSockd) Broadcast(roomName string, message []byte) error {
	room := s.getRoom(roomName, false)
	if room == nil {
		return nil
	}

	return room.Broadcast(message)
}

func (s *HigherSockd) Publish(roomName string, topicName string, message []byte) error {
	room := s.getRoom(roomName, false)
	if room == nil {
		return errors.New("room not found")
	}

	return room.Publish(topicName, message)

}

// DirectMessage sends a message to a specific user
func (s *HigherSockd) DirectMessage(roomName string, targetConnId int64, message []byte) error {
	room := s.getRoom(roomName, false)
	if room == nil {
		return nil
	}

	return room.DirectMessage(targetConnId, message)
}

// Subscribe adds a connection to a topic subscription
func (s *HigherSockd) Subscribe(roomName string, topicName string, connId int64) error {
	room := s.getRoom(roomName, false)
	if room == nil {
		return errors.New("room not found")
	}

	return room.Subscribe(topicName, connId)
}

// Unsubscribe removes a connection from a topic subscription
func (s *HigherSockd) Unsubscribe(roomName string, topicName string, connId int64) error {
	room := s.getRoom(roomName, false)
	if room == nil {
		return errors.New("room not found")
	}

	return room.Unsubscribe(topicName, connId)

}
