package broadcast

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
	broadcast  chan []byte

	// sessions: ConnId -> Session Object
	sessions map[int64]*session
	sLock    sync.RWMutex
}

func (r *Room) AddConn(userId int64, conn net.Conn, connId int64) (int64, error) {
	sess := &session{
		room:   r,
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
	go sess.readPump()

	if existingSess != nil {
		existingSess.teardown()
	}

	return sess.connId, nil
}

func (r *Room) RemoveConn(userId int64, connId int64) error {

	tcan := time.After(time.Second * 10)

	select {
	case r.disconnect <- connId:
		return nil
	case <-tcan:
		return errors.New("room is very busy or dead")
	}

}

func (r *Room) Broadcast(message []byte) error {

	sessions := make([]*session, 0, len(r.sessions))

	r.sLock.RLock()
	for _, sess := range r.sessions {
		sessions = append(sessions, sess)
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

// private

func (r *Room) run() {

	for {
		select {
		case msg := <-r.broadcast:
			copySess := make([]*session, 0, len(r.sessions))

			r.sLock.RLock()
			for _, sess := range r.sessions {
				copySess = append(copySess, sess)
			}
			r.sLock.RUnlock()

			for _, sess := range copySess {

				tcan := time.After(time.Second * 1)

				select {
				case sess.send <- msg:
					continue
				case <-tcan:
					qq.Println("@drop_message", sess.connId)
					continue
				}

			}

		case connId := <-r.disconnect:
			r.cleanup(connId)
		}
	}

}

func (r *Room) cleanup(connId int64) {

	r.sLock.Lock()
	sess, exists := r.sessions[connId]
	delete(r.sessions, connId)
	r.sLock.Unlock()

	if !exists {
		return
	}

	sess.teardown()

}
