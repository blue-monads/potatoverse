package sockd

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blue-monads/turnix/backend/utils/qq"
)

type Sockd struct {
	rooms map[string]*Room
	mu    sync.RWMutex

	counter atomic.Int32
}

func NewSockd() *Sockd {
	return &Sockd{
		rooms: make(map[string]*Room),
	}
}

func newRoom(name string) *Room {
	r := &Room{
		name:       name,
		disconnect: make(chan int32, 32), // Buffer for burst disconnects
		topics:     make(map[string]map[int32]bool),
		sessions:   make(map[int32]*session),
	}

	// Start the Room Event Loop
	go r.run()

	return r
}

func (s *Sockd) AddConn(userId int64, conn net.Conn, roomName string) (int32, error) {
	s.mu.Lock()
	room, exists := s.rooms[roomName]
	if !exists {
		room = newRoom(roomName)
		s.rooms[roomName] = room
	}
	s.mu.Unlock()

	sess := &session{
		room:   room, // Link back to room
		connId: s.counter.Add(1),
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

	return sess.connId, nil
}

func (s *Sockd) RemoveConn(userId int64, connId int32, roomName string) error {
	s.mu.RLock()
	room, exists := s.rooms[roomName]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	// We simply push to the disconnect channel.
	// The event loop (run) handles the actual locking and map deletion.
	// This keeps logic centralized and prevents race conditions between
	// manual removal and network-error removal.
	select {
	case room.disconnect <- connId:
		// Signal sent
	default:
		time.Sleep(time.Second * 2)
		select {
		case room.disconnect <- connId:
			// Signal sent
		default:
			return errors.New("room is very busy or dead")
		}
	}

	return nil
}

func (s *Sockd) Publish(roomName string, topicName string, message []byte) error {
	s.mu.RLock()
	room, exists := s.rooms[roomName]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	// Get Subscribers
	room.tLock.RLock()
	subMap, ok := room.topics[topicName]
	if !ok || len(subMap) == 0 {
		room.tLock.RUnlock()
		return nil
	}

	// Snapshot IDs
	ids := make([]int32, 0, len(subMap))
	for id := range subMap {
		ids = append(ids, id)
	}
	room.tLock.RUnlock()

	// Send

	room.sLock.RLock()

	sessions := make([]*session, 0, len(subMap))
	for _, id := range ids {
		if sess, ok := room.sessions[id]; ok {
			sessions = append(sessions, sess)
		}
	}

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

	room.sLock.RUnlock()

	return nil
}

func (s *Sockd) AddSub(roomName string, topicName string, userId int64, connId int32, conn net.Conn) error {
	s.mu.RLock()
	room, exists := s.rooms[roomName]
	s.mu.RUnlock()
	if !exists {
		return errors.New("room not found")
	}

	room.tLock.Lock()
	if room.topics[topicName] == nil {
		room.topics[topicName] = make(map[int32]bool)
	}
	room.topics[topicName][connId] = true
	room.tLock.Unlock()
	return nil
}
