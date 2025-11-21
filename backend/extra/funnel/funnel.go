package funnel

import (
	"net"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// Funnel is a service that routes all http requests to a node(server) which are connected
// to the service through websocket becase the service is not accessible from the internet (behind NAT)

type ServerConnection struct {
	Conn net.Conn
}

type PendingRequest struct {
	ResponseChan chan *http.Response
	ErrorChan    chan error
}

type Funnel struct {
	serverConnections map[string]*ServerConnection
	scLock            sync.RWMutex

	pendingRequests map[string]*PendingRequest
	prLock          sync.RWMutex
}

// New creates a new Funnel instance
func New() *Funnel {
	return &Funnel{
		serverConnections: make(map[string]*ServerConnection),
		pendingRequests:   make(map[string]*PendingRequest),
	}
}

func (f *Funnel) HandleServerWebSocket(serverId string, c *gin.Context) {
	f.handleServerWebSocket(serverId, c)
}

func (f *Funnel) HandleRoute(serverId string, c *gin.Context) {

	if c.Request.Header.Get("Upgrade") == "websocket" {
		f.routeWS(serverId, c)
		return
	}

	f.routeHttp(serverId, c)
}
