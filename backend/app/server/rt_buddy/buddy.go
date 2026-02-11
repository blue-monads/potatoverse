package rtbuddy

import (
	"sync"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/buddyhub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/selfcdc"
	"github.com/gin-gonic/gin"
)

const (
	BuddyAuthExpiry = 5 * time.Minute
)

type BuddyRouteServer struct {
	buddyhub     *buddyhub.BuddyHub
	port         int
	serverPubKey string

	// lazy cdc
	selfcdc *selfcdc.SelfCDCSyncer

	allowAnyBuddy bool

	reverseBuddyIdToPubkey map[string]string
	rLock                  sync.RWMutex
}

func New(buddyhub *buddyhub.BuddyHub, port int, serverPubKey string) *BuddyRouteServer {
	return &BuddyRouteServer{
		buddyhub:               buddyhub,
		port:                   port,
		serverPubKey:           serverPubKey,
		reverseBuddyIdToPubkey: map[string]string{},
		rLock:                  sync.RWMutex{},
		allowAnyBuddy:          true,
	}
}

func (a *BuddyRouteServer) AttachRoutes(g *gin.RouterGroup) {
	g.POST("/buddy/ping", a.handleBuddyPing)
	g.Any("/buddy/route", a.handleBuddyRoute)
	g.GET("/buddy/register", a.registerBuddyNode)

	// lazysync
	g.POST("/buddy/lazycdc/sync/data", a.handleBuddyLazySyncData)
	g.GET("/buddy/lazycdc/sync/meta", a.handleBuddyLazySyncMeta)
}

// npub19z2svstd8aega8rtld86lc5wdjmgz0jnwku9u62mcdrl60ge2pespjwl6j
func (b *BuddyRouteServer) setNode(pubkey string) {
	// Safety check
	if len(pubkey) < 20 {
		panic("pubkey is too short")
	}

	firstId := pubkey[5:12]
	lastId := pubkey[len(pubkey)-7:]

	b.rLock.Lock()
	defer b.rLock.Unlock()

	b.reverseBuddyIdToPubkey[firstId] = pubkey
	b.reverseBuddyIdToPubkey[lastId] = pubkey
	b.reverseBuddyIdToPubkey[pubkey] = pubkey
}

func (b *BuddyRouteServer) getNodeId(id string) string {
	b.rLock.RLock()
	defer b.rLock.Unlock()

	return b.reverseBuddyIdToPubkey[id]

}
