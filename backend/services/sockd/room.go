package sockd

import "sync"

type Room struct {
	name string

	disconnect chan int32

	topics map[string]map[int32]bool
	tLock  sync.RWMutex

	// sessions: ConnId -> Session Object
	sessions map[int32]*session
	sLock    sync.RWMutex
}

func (r *Room) run() {
	for connId := range r.disconnect {
		r.cleanup(connId)
	}
}

// cleanup performs the heavy lifting of removing the user from all maps
func (r *Room) cleanup(connId int32) {
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
