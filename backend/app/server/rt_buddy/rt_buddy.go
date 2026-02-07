package rtbuddy

import (
	"net/http"
	"sync"
	"time"

	"github.com/blue-monads/potatoverse/backend/app/server/rt_buddy/webdav"
	"github.com/blue-monads/potatoverse/backend/services/buddyhub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/selfcdc"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/gin-gonic/gin"
)

const (
	BuddyAuthExpiry = 5 * time.Minute
)

type BuddyRouteServer struct {
	buddyhub      *buddyhub.BuddyHub
	port          int
	serverPubKey  string
	webdavServers map[string]*webdav.WebdavServer
	webdavLock    sync.RWMutex

	// lazy cdc
	selfcdc *selfcdc.SelfCDCSyncer
}

func New(buddyhub *buddyhub.BuddyHub, port int, serverPubKey string) *BuddyRouteServer {
	return &BuddyRouteServer{
		buddyhub:      buddyhub,
		port:          port,
		serverPubKey:  serverPubKey,
		webdavServers: make(map[string]*webdav.WebdavServer),
		webdavLock:    sync.RWMutex{},
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

func (a *BuddyRouteServer) handleBuddyPing(ctx *gin.Context) {

	pubkey, err := verifyNostrAuthCtx(ctx, BuddyAuthExpiry)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	qq.Println("@buddy_pubkey", pubkey)

	serverPubkey := a.buddyhub.GetPubkey()

	ctx.JSON(http.StatusOK, gin.H{
		"message":       "Buddy pinged",
		"server_pubkey": serverPubkey,
	})

}
