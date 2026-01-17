package buddyhub

import (
	"net/http"
	"sync"

	"github.com/blue-monads/potatoverse/backend/xtypes"
)

type ProviderHandle struct {
	Pubkey             string
	URL                string
	Provider           string
	App                xtypes.App
	SetConnectionState func(bpubkey string, state bool) error
}

type WSRequest struct {
	URL     string
	Inchan  chan []byte
	Outchan chan []byte
}

type BuddyConnection interface {
	Ping() (bool, error)
	PerformRequest(buddyPubkey string, request *http.Request) (*http.Response, error)
	PerformWebSocketRequest(buddyPubkey string, wsrequest *WSRequest) error
	Close() error
}

type BuddyConnectionContainer struct {
	connections map[string][]string
	mu          sync.RWMutex
}
