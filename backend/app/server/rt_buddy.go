package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/gin-gonic/gin"
	"github.com/nbd-wtf/go-nostr"
)

const (
	BuddyAuthExpiry = 5 * time.Minute
)

func verifyNostrAuth(ctx *gin.Context, expiry time.Duration) (*nostr.Event, error) {
	authHeader := ctx.GetHeader("X-Buddy-Auth")
	if authHeader == "" {
		return nil, fmt.Errorf("Unauthorized")
	}

	eventJson, err := base64.StdEncoding.DecodeString(authHeader)
	if err != nil {
		return nil, fmt.Errorf("Invalid authorization header")
	}

	var event nostr.Event
	err = json.Unmarshal(eventJson, &event)
	if err != nil {
		return nil, fmt.Errorf("Invalid authorization header")
	}

	ok, err := event.CheckSignature()
	if !ok || err != nil {
		return nil, fmt.Errorf("invalid signature")
	}

	if event.Kind != nostr.KindHTTPAuth {
		return nil, fmt.Errorf("wrong event kind")
	}

	if time.Since(event.CreatedAt.Time()) > expiry {
		return nil, fmt.Errorf("event expired")
	}

	return &event, nil
}

func (a *Server) handleBuddyPing(ctx *gin.Context) {

	pubkey, err := verifyNostrAuth(ctx, BuddyAuthExpiry)
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

func (a *Server) handleBuddyRoute(ctx *gin.Context) {
	ev, err := verifyNostrAuth(ctx, BuddyAuthExpiry)
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

	newHost := fmt.Sprintf("localhost:%d", a.opt.Port)

	if strings.HasPrefix(host, "zz-") {
		parts := strings.Split(host, ".")
		suborigin := parts[len(parts)-1]
		newHost = fmt.Sprintf("%s.localhost:%d", suborigin, a.opt.Port)
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
