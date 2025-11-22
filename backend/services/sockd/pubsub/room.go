package pubsub

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/blue-monads/turnix/backend/utils/qq"
)

type Room struct {
	name string

	disconnect chan int64

	topics map[string]map[int64]bool
	tLock  sync.RWMutex

	// sessions: ConnId -> Session Object
	sessions map[int64]*session
	sLock    sync.RWMutex
}

func (r *Room) run() {
	for connId := range r.disconnect {
		r.cleanup(connId)
	}
}

func (r *Room) AddConn(userId int64, conn net.Conn, connId int64) (int64, error) {

	sess := &session{
		room:   r, // Link back to room
		connId: connId,
		userId: userId,
		conn:   conn,
		send:   make(chan []byte, 16),
	}

	r.sLock.Lock()
	existingSess := r.sessions[sess.connId]
	r.sessions[sess.connId] = sess
	r.sLock.Unlock()

	go sess.writePump()

	if existingSess != nil {
		existingSess.teardown()
	}

	return sess.connId, nil
}

func (r *Room) RemoveConn(userId int64, connId int64) error {

	// We simply push to the disconnect channel.
	// The event loop (run) handles the actual locking and map deletion.
	// This keeps logic centralized and prevents race conditions between
	// manual removal and network-error removal.
	select {
	case r.disconnect <- connId:
		// Signal sent
	default:
		time.Sleep(time.Second * 2)
		select {
		case r.disconnect <- connId:
			// Signal sent
		default:
			return errors.New("room is very busy or dead")
		}
	}

	return nil
}

func (r *Room) Publish(topicName string, message []byte) error {

	// Get Subscribers
	r.tLock.RLock()
	subMap, ok := r.topics[topicName]
	if !ok || len(subMap) == 0 {
		r.tLock.RUnlock()
		return nil
	}

	// Snapshot IDs
	ids := make([]int64, 0, len(subMap))
	for id := range subMap {
		ids = append(ids, id)
	}
	r.tLock.RUnlock()

	// Send

	r.sLock.RLock()
	sessions := make([]*session, 0, len(subMap))
	for _, id := range ids {
		if sess, ok := r.sessions[id]; ok {
			sessions = append(sessions, sess)
		}
	}
	r.sLock.RUnlock()

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

func (r *Room) AddSub(topicName string, userId int64, connId int64) error {
	r.tLock.Lock()
	if r.topics[topicName] == nil {
		r.topics[topicName] = make(map[int64]bool)
	}
	r.topics[topicName][connId] = true
	r.tLock.Unlock()
	return nil
}

// private

// cleanup performs the heavy lifting of removing the user from all maps
func (r *Room) cleanup(connId int64) {
	// 1. Remove from Session Map
	r.sLock.Lock()
	sess, exists := r.sessions[connId]
	if !exists {
		r.sLock.Unlock()
		return // Already cleaned up
	}
	delete(r.sessions, connId)
	r.sLock.Unlock()

	sess.teardown()

	r.tLock.Lock()
	for topicName, subscribers := range r.topics {
		if _, ok := subscribers[connId]; ok {
			delete(subscribers, connId)
			// Optional: Delete empty topics
			if len(subscribers) == 0 {
				delete(r.topics, topicName)
			}
		}
	}
	r.tLock.Unlock()
}
