package notifier

import (
	"net"
	"sync"

	"github.com/blue-monads/turnix/backend/services/datahub"
)

type Notifier struct {
	userConnections map[int64]*UserRoom
	mu              sync.RWMutex
	database        datahub.UserOps
	maxMsgId        int64
}

type UserRoom struct {
	userId      int64
	group       string
	maxMsgId    int64
	connections map[string]Connection
	mu          sync.RWMutex
}

type Connection struct {
	connId int64
	conn   net.Conn
}
