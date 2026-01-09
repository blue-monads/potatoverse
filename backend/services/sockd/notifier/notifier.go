package notifier

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
)

type Notifier struct {
	userConnections map[int64]*UserRoom
	mu              sync.RWMutex
	maxMsgId        int64

	connIdCounter atomic.Int64

	cleanConnChan chan int64
}

func (n *Notifier) Run() {
	for connId := range n.cleanConnChan {
		room := n.getUserRoom(connId)
		if room != nil {
			room.performCleanupConn(connId)
		}
	}
}

// New creates a new Notifier instance
func New() Notifier {

	return Notifier{
		userConnections: make(map[int64]*UserRoom),
		maxMsgId:        0,
		connIdCounter:   atomic.Int64{},
	}
}

// getUserRoom gets or creates a UserRoom for a user
func (n *Notifier) getUserRoom(userId int64) *UserRoom {
	n.mu.RLock()
	room := n.userConnections[userId]
	n.mu.RUnlock()

	return room
}

func (n *Notifier) getUserRoomOrCreate(userId int64, group string) *UserRoom {
	n.mu.RLock()
	room, exists := n.userConnections[userId]
	n.mu.RUnlock()

	if !exists {
		n.mu.Lock()
		// Double-check after acquiring write lock
		room, exists = n.userConnections[userId]
		if !exists {
			room = &UserRoom{
				userId:      userId,
				group:       group,
				maxMsgId:    0,
				connections: make(map[int64]*Connection),
			}
			n.userConnections[userId] = room
		}
		n.mu.Unlock()
	}

	return room
}

func (n *Notifier) AddUserConnection(userId int64, groupName string, conn net.Conn) (int64, error) {
	connId := n.connIdCounter.Add(1)

	room := n.getUserRoomOrCreate(userId, groupName)
	if room == nil {
		return 0, errors.New("failed to get or create user room")
	}

	return room.AddUserConnection(connId, conn)
}

func (n *Notifier) RemoveUserConnection(userId int64, connId int64) error {
	room := n.getUserRoom(userId)
	if room == nil {
		return nil
	}

	return room.RemoveUserConnection(connId)

}

func (n *Notifier) SendUserMessage(userId int64, msg *dbmodels.UserMessage) error {

	room := n.getUserRoom(userId)
	if room == nil {
		return nil // User has no connections
	}

	return room.SendUserMessage(msg)

}

func (n *Notifier) SendUser(userId int64, message []byte) error {
	room := n.getUserRoom(userId)
	if room == nil {
		return nil // User has no connections
	}

	return room.SendUser(message)
}

func (n *Notifier) BroadcastGroup(groupName string, message []byte) error {

	n.mu.RLock()
	rooms := make([]*UserRoom, 0)
	for _, room := range n.userConnections {
		if room.group == groupName {
			rooms = append(rooms, room)
		}
	}
	n.mu.RUnlock()

	for _, room := range rooms {
		room.mu.RLock()
		connections := make([]*Connection, 0, len(room.connections))
		for _, conn := range room.connections {
			connections = append(connections, conn)
		}
		room.mu.RUnlock()

		for _, conn := range connections {
			select {
			case conn.send <- message:
			case <-time.After(time.Second * 5):
				qq.Println("@BroadcastGroup/timeout", conn.connId)
			}
		}
	}

	return nil
}

func (n *Notifier) BroadcastAll(message []byte) error {

	n.mu.RLock()
	rooms := make([]*UserRoom, 0, len(n.userConnections))
	for _, room := range n.userConnections {
		rooms = append(rooms, room)
	}
	n.mu.RUnlock()

	for _, room := range rooms {
		room.mu.RLock()
		connections := make([]*Connection, 0, len(room.connections))
		for _, conn := range room.connections {
			connections = append(connections, conn)
		}
		room.mu.RUnlock()

		for _, conn := range connections {
			select {
			case conn.send <- message:
			case <-time.After(time.Second * 5):
				qq.Println("@BroadcastAll/timeout", conn.connId)
			}
		}
	}

	return nil
}
