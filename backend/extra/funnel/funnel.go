package funnel

import (
	"net"
	"sync"

	"github.com/gin-gonic/gin"
)

type ServerConnection struct {
	Conn net.Conn
}

type Funnel struct {
	ServerConnections map[string]*ServerConnection
	scLock            sync.RWMutex
}

func (f *Funnel) HandleConnect(c *gin.Context) {}
