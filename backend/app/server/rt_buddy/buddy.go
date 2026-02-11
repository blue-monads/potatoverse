package rtbuddy

import (
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
}

func New(buddyhub *buddyhub.BuddyHub, port int, serverPubKey string) *BuddyRouteServer {
	return &BuddyRouteServer{
		buddyhub:     buddyhub,
		port:         port,
		serverPubKey: serverPubKey,
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
