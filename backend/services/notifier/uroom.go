package notifier

import "sync"

type UserRoom struct {
	notifier    *Notifier
	userId      int64
	group       string
	maxMsgId    int64
	connections map[int64]*Connection
	mu          sync.RWMutex
}

func (u *UserRoom) performCleanupConn(connId int64) {

	u.mu.Lock()
	defer u.mu.Unlock()

	conn, exists := u.connections[connId]
	if !exists {
		return
	}

	conn.teardown()
	delete(u.connections, connId)

}
