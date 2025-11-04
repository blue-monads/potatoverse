package funnel

import (
	"net"
	"sync"

	"github.com/gin-gonic/gin"
)

// root_<pub_key_hash>.freehttptunnel.com
// <s-x>_<pub_key_hash>.freehttptunnel.com

type ServerConnection struct {
	Conn net.Conn
}

type Funnel struct {
	ServerConnections map[string]*ServerConnection
	scLock            sync.RWMutex
}

func (f *Funnel) HandleVerify(c *gin.Context) {}

func (f *Funnel) HandleConnect(c *gin.Context) {}

func SignWithPubKey(pubKey, signPayload string) (string, error) {

	return "", nil
}
