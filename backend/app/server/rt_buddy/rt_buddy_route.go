package rtbuddy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/blue-monads/potatoverse/backend/utils/nostrutils"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/publicsuffix"
)

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

func (a *BuddyRouteServer) BuddyAutoRouteMW() gin.HandlerFunc {
	pubkey := a.buddyhub.GetPubkey()

	return func(ctx *gin.Context) {

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

		if strings.HasPrefix(subdomain, "buddy") {
			a.routeToBuddy(subdomain, ctx)
			return
		}

		if strings.HasPrefix(subdomain, "zz-") && strings.Contains(subdomain, "buddy") {
			a.routeToBuddy(subdomain, ctx)
			return
		}
	}

	// buddy end

}

func (a *BuddyRouteServer) routeToBuddy(subdomain string, ctx *gin.Context) {

	extractedPubkey := strings.Split(subdomain, "buddy")[1]
	a.buddyhub.HandleFunnelRoute(fmt.Sprintf("npub%s", extractedPubkey), ctx)
}

func (a *BuddyRouteServer) registerBuddyNode(ctx *gin.Context) {

	token := ctx.Query("token")

	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	pubkey, err := nostrutils.DecodeKeyToHex(token)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	a.buddyhub.HandleFunnelRegisterNode(fmt.Sprintf("npub%s", pubkey), ctx)

}

// private

func getSubdomain(host string) (string, error) {

	if before, ok := strings.CutSuffix(host, ".localhost"); ok {
		qq.Println("@getSubdomain/0", before)
		return before, nil
	}

	// 2. Get the Registered Domain (e.g., "example.co.uk")
	mainDomain, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return "", err
	}

	// 3. Remove the main domain from the host to get the subdomain
	subdomain := strings.TrimSuffix(host, mainDomain)
	subdomain = strings.TrimSuffix(subdomain, ".") // Remove trailing dot

	qq.Println("@getSubdomain/1", host, mainDomain, subdomain)

	if strings.Contains(subdomain, ".") {
		parts := strings.Split(subdomain, ".")
		subdomain = parts[0]
	}

	qq.Println("@getSubdomain/2", subdomain)

	return subdomain, nil
}
