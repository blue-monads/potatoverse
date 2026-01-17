package notifier

import (
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
)

type UserRoom struct {
	notifier    *Notifier
	userId      int64
	group       string
	maxMsgId    int64
	connections map[int64]*Connection
	mu          sync.RWMutex
}

func (u *UserRoom) AddUserConnection(connId int64, conn net.Conn) (int64, error) {

	connection := &Connection{
		connId:           connId,
		conn:             conn,
		send:             make(chan []byte, 16),
		closedAndCleaned: false,
		userRoom:         u,
	}

	u.mu.Lock()
	existingConn := u.connections[connId]
	u.connections[connId] = connection
	u.mu.Unlock()

	go connection.writePump()

	if existingConn != nil {
		existingConn.teardown()
	}

	return connId, nil
}

func (u *UserRoom) RemoveUserConnection(connId int64) error {

	u.mu.Lock()
	conn, exists := u.connections[connId]
	if exists {
		delete(u.connections, connId)
	}
	u.mu.Unlock()

	if exists {
		conn.teardown()
	}

	// Clean up empty rooms
	u.mu.RLock()
	isEmpty := len(u.connections) == 0
	u.mu.RUnlock()

	if isEmpty {
		u.notifier.mu.Lock()
		delete(u.notifier.userConnections, u.userId)
		u.notifier.mu.Unlock()
	}

	return nil
}

func (u *UserRoom) SendUserMessage(msg *dbmodels.UserMessage) error {
	message, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return u.SendUser(message)
}

func (u *UserRoom) SendUser(messageBytes []byte) error {

	u.mu.RLock()
	connections := make([]*Connection, 0, len(u.connections))
	for _, conn := range u.connections {
		connections = append(connections, conn)
	}
	u.mu.RUnlock()

	for _, conn := range connections {
		select {
		case conn.send <- messageBytes:
		case <-time.After(time.Second * 5):
			qq.Println("@SendUser/timeout", conn.connId)
		}
	}

	return nil
}

// private

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
