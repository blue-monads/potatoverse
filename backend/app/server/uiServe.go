package server

import (
	"fmt"
	"io"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/blue-monads/turnix/backend/app/server/assets"
	"github.com/blue-monads/turnix/frontend"

	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
)

var (
	DEV_MODE = true
)

// during dev we just proxy to dev vite server running otherwise serve files from build folder
func (s *Server) pages(z *gin.RouterGroup) {
	rfunc := assets.PagesRoutesServer()
	extFunc := s.externalAssets()

	z.GET("/pages", rfunc)
	z.GET("/pages/*files", rfunc)
	z.GET("/spaces/:pname/*files", extFunc)
	z.GET("/spaces/:pname", extFunc)
	z.GET("/lib/*file", func(ctx *gin.Context) {

		pp.Println("@lib/1")

		file := ctx.Param("file")

		file = strings.TrimPrefix(file, "/")

		fout, err := frontend.BuildProd.Open(path.Join("output", file))
		if err != nil {
			pp.Println("@lib/2", err.Error())
			return
		}

		pp.Println("@lib/3")

		defer fout.Close()

		io.Copy(ctx.Writer, fout)

	})

}

func (s *Server) externalAssets() gin.HandlerFunc {

	proxyAddrs := map[string]*httputil.ReverseProxy{}

	if DEV_MODE {

		pp.Println("@ext/0")

		devSpacesEnv := os.Getenv("TURNIX_DEV_SPACES")
		devSpaces := strings.Split(devSpacesEnv, ",")

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

		pp.Println("@ext/1")

		pname := ctx.Param("pname")

		if DEV_MODE {
			pp.Println("@ext/2")
			proxy := proxyAddrs[pname]
			if proxy != nil {
				pp.Println("@ext/3")
				proxy.ServeHTTP(ctx.Writer, ctx.Request)
				return
			}
			pp.Println("@ext/4")
		}

	}

}
