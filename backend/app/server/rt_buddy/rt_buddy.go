package rtbuddy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/blue-monads/turnix/backend/app/server/rt_buddy/webdav"
	"github.com/blue-monads/turnix/backend/services/corehub/buddyhub"
	"github.com/blue-monads/turnix/backend/utils/qq"
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
	g.Any("/buddy/webdav/*path", a.handleBuddyWebdav)
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

func (a *BuddyRouteServer) handleBuddyRoute(ctx *gin.Context) {
	ev, err := verifyNostrAuthCtx(ctx, BuddyAuthExpiry)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	turl := ev.Tags[1][1]

	// convert https://example.com/ to http://localhost:3000/
	// convert zz-12-serverkey.example.com to http://zz-12-serverkey.localhost:3000/

	purl, err := url.Parse(turl)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	host := purl.Host

	newHost := fmt.Sprintf("localhost:%d", a.port)

	if strings.HasPrefix(host, "zz-") {
		parts := strings.Split(host, ".")
		suborigin := parts[len(parts)-1]
		newHost = fmt.Sprintf("%s.localhost:%d", suborigin, a.port)
	}

	newUrl := url.URL{
		Scheme:   "http",
		Host:     newHost,
		Path:     purl.Path,
		RawQuery: purl.RawQuery,
		Fragment: purl.Fragment,
	}

	proxy := httputil.NewSingleHostReverseProxy(&newUrl)
	proxy.ServeHTTP(ctx.Writer, ctx.Request)

}

func (a *BuddyRouteServer) BuddyAutoRouteMW(ctx *gin.Context) {

	pubkey := a.buddyhub.GetPubkey()

	domainName := ctx.Request.Host
	if strings.Contains(domainName, ":") {
		hh, _, err := net.SplitHostPort(ctx.Request.Host)
		if err != nil {
			domainName = ctx.Request.Host
		} else {
			domainName = hh
		}
	}

	qq.Println("@BuddyAutoRouteMW/1", domainName)

	subdomain, err := getSubdomain(domainName)
	if err != nil {
		return
	}

	// current node start
	if subdomain == "" || subdomain == "main" || subdomain == pubkey {
		ctx.Next()
		return
	}

	if strings.HasPrefix(subdomain, "zz-") && strings.HasSuffix(subdomain, "-main") {
		ctx.Next()
		return
	}

	if strings.HasPrefix(subdomain, "zz-") && strings.HasSuffix(subdomain, a.serverPubKey) {
		ctx.Next()
		return
	}

	// current node end

	// buddy start

	if strings.HasPrefix(subdomain, "npub") {
		a.routeToBuddy(subdomain, ctx)
		return
	}

	if strings.HasPrefix(subdomain, "zz-") && strings.Contains(subdomain, "npub") {
		a.routeToBuddy(subdomain, ctx)
		return
	}

	// buddy end

}

func (a *BuddyRouteServer) routeToBuddy(subdomain string, ctx *gin.Context) {
	extractedPubkey := strings.Split(subdomain, "npub")[1]
	a.buddyhub.HandleFunnelRoute(fmt.Sprintf("npub%s", extractedPubkey), ctx)
}
