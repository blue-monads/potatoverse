package funnel

import (
	"net"
	"sync"
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
