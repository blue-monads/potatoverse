package broadcast

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/blue-monads/turnix/backend/utils/qq"
)

type BroadcastSockd struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

func NewSockd() BroadcastSockd {
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

func (s *BroadcastSockd) AddConn(userId int64, conn net.Conn, connId int64, roomName string) (int64, error) {
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
	if _, ok := room.sessions[sess.connId]; ok {
		room.sLock.Unlock()
		return 0, errors.New("connId collision")
	}
	room.sessions[sess.connId] = sess
	room.sLock.Unlock()

	go sess.writePump()
	go sess.readPump()

	return sess.connId, nil
}

func (s *BroadcastSockd) RemoveConn(userId int64, connId int64, roomName string) error {
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

func (s *BroadcastSockd) Broadcast(roomName string, message []byte) error {
	s.mu.RLock()
	room, exists := s.rooms[roomName]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	sessions := make([]*session, 0, len(room.sessions))

	room.sLock.RLock()
	for _, sess := range room.sessions {
		sessions = append(sessions, sess)
	}
	room.sLock.RUnlock()

	for _, sess := range sessions {

		tcan := time.After(time.Second * 5)

		select {
		case sess.send <- message:
			continue
		case <-tcan:
			qq.Println("@publish/timeout", sess.connId)
			continue
		}
	}

	return nil
}
