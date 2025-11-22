package pubsub

import (
	"errors"
	"net"
	"sync"
)

type PubSubSockd struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func NewSockd() PubSubSockd {
	return PubSubSockd{
		rooms: make(map[string]*Room),
	}
}

func newRoom(name string) *Room {
	r := &Room{
		name:       name,
		disconnect: make(chan int64, 32), // Buffer for burst disconnects
		topics:     make(map[string]map[int64]bool),
		sessions:   make(map[int64]*session),
	}

	// Start the Room Event Loop
	go r.run()

	return r
}

func (s *PubSubSockd) getRoom(roomName string, createIfNotExists bool) *Room {
	s.mu.RLock()
	room, exists := s.rooms[roomName]
	s.mu.RUnlock()
	if !exists {
		if createIfNotExists {
			room = newRoom(roomName)
			s.mu.Lock()
			defer s.mu.Unlock()
			sneakyRoom := s.rooms[roomName]
			if sneakyRoom != nil {
				return sneakyRoom
			}
			s.rooms[roomName] = room
			go room.run()
		}
		return room
	}
	return room
}

func (s *PubSubSockd) AddConn(userId int64, conn net.Conn, connId int64, roomName string) (int64, error) {
	room := s.getRoom(roomName, true)
	if room == nil {
		return 0, errors.New("room not found")
	}

	return room.AddConn(userId, conn, connId)

}

func (s *PubSubSockd) RemoveConn(userId int64, connId int64, roomName string) error {
	room := s.getRoom(roomName, false)
	if room == nil {
		return nil
	}

	return room.RemoveConn(userId, connId)

}

func (s *PubSubSockd) Publish(roomName string, topicName string, message []byte) error {
	room := s.getRoom(roomName, false)
	if room == nil {
		return nil
	}

	return room.Publish(topicName, message)

}

func (s *PubSubSockd) AddSub(roomName string, topicName string, userId int64, connId int64, conn net.Conn) error {
	room := s.getRoom(roomName, false)
	if room == nil {
		return errors.New("room not found")
	}

	return room.AddSub(topicName, userId, connId)

}
