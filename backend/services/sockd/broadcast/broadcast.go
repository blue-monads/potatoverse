package broadcast

import (
	"errors"
	"net"
	"sync"
)

type BroadcastSockd struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func New() BroadcastSockd {
	return BroadcastSockd{
		rooms: make(map[string]*Room),
		mu:    sync.RWMutex{},
	}
}

func newRoom(name string) *Room {
	r := &Room{
		name:       name,
		disconnect: make(chan int64, 32), // Buffer for burst disconnects
		sessions:   make(map[int64]*session),
		broadcast:  make(chan []byte, 32),
		sLock:      sync.RWMutex{},
	}

	// Start the Room Event Loop
	go r.run()

	return r
}

func (s *BroadcastSockd) getRoom(roomName string, createIfNotExists bool) *Room {
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

func (s *BroadcastSockd) AddConn(userId int64, conn net.Conn, connId int64, roomName string) (int64, error) {
	room := s.getRoom(roomName, true)
	if room == nil {
		return 0, errors.New("room not found")
	}

	return room.AddConn(userId, conn, connId)
}

func (s *BroadcastSockd) RemoveConn(userId int64, connId int64, roomName string) error {

	room := s.getRoom(roomName, false)
	if room == nil {
		return nil
	}

	return room.RemoveConn(userId, connId)
}

func (s *BroadcastSockd) Broadcast(roomName string, message []byte) error {

	room := s.getRoom(roomName, false)
	if room == nil {
		return nil
	}

	return room.Broadcast(message)

}
