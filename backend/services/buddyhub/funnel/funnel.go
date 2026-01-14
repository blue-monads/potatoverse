package funnel

import (
	"net"
	"sync"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/gin-gonic/gin"
)

// Funnel is a service that routes all http requests to a node(server) which are connected
// to the service through websocket becase the service is not accessible from the internet (behind NAT)

type ServerHandle struct {
	conn      net.Conn
	writeChan chan *ServerWrite
}

type ServerWrite struct {
	packet *Packet
	reqId  string
}

type Funnel struct {
	serverConnections map[string]*ServerHandle
	scLock            sync.RWMutex

	pendingReq     map[string]chan *Packet
	pendingReqLock sync.Mutex
}

// New creates a new Funnel instance
func New() *Funnel {
	return &Funnel{
		serverConnections: make(map[string]*ServerHandle),
		scLock:            sync.RWMutex{},
		pendingReq:        make(map[string]chan *Packet),
		pendingReqLock:    sync.Mutex{},
	}
}

func (f *Funnel) HandleServerWebSocket(serverId string, c *gin.Context) {
	qq.Println("@Funnel/HandleServerWebSocket/1{SERVER_ID}", serverId)

	f.handleServerWebSocket(serverId, c)
}

func (f *Funnel) HandleRoute(serverId string, c *gin.Context) {

	if c.Request.Header.Get("Upgrade") == "websocket" {
		f.routeWS(serverId, c)
		return
	}

	f.routeHttp(serverId, c)
}
