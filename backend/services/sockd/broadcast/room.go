package broadcast

import (
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
