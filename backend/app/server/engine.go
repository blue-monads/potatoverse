package server

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
)

type InstallPackageRequest struct {
	URL string `json:"url"`
}

func (a *Server) InstallPackage(ctx *gin.Context) {
	var req InstallPackageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	packageId, err := a.engine.InstallPackageByUrl(req.URL)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"package_id": packageId})

}

func (a *Server) handleSpaceFile() func(ctx *gin.Context) {

	proxyAddrs := map[string]*httputil.ReverseProxy{}

	if DEV_MODE {
		devSpacesEnv := os.Getenv("TURNIX_DEV_SPACES")
		devSpaces := strings.Split(devSpacesEnv, ",")

		pp.Println("@devSpaces", devSpaces)

		for _, pname := range devSpaces {
			nameParts := strings.Split(pname, ":")
			if len(nameParts) != 2 {
				continue
			}

			url, err := url.Parse(fmt.Sprint("http://localhost:", nameParts[1]))
			if err != nil {
				panic(err)
			}
			proxy := httputil.NewSingleHostReverseProxy(url)
			proxyAddrs[nameParts[0]] = proxy
		}
	}

	return func(ctx *gin.Context) {
		spaceKey := ctx.Param("space_key")
		proxy := proxyAddrs[spaceKey]
		if proxy != nil {
			proxy.ServeHTTP(ctx.Writer, ctx.Request)
			return
		}

		a.engine.ServeSpaceFile(ctx)

	}
}

func (a *Server) handlePluginFile() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {}
}

func (a *Server) handleSpaceApi(ctx *gin.Context) {}

func (a *Server) handlePluginApi(ctx *gin.Context) {}
