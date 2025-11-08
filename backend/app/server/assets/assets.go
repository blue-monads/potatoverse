package assets

import (
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/utils/qq"
	output "github.com/blue-monads/turnix/frontend"
	"github.com/gin-gonic/gin"
)

const NoPreBuildFiles = false

func PagesRoutesServer() gin.HandlerFunc {

	var proxy *httputil.ReverseProxy
	pserver := os.Getenv("FRONTEND_DEV_SERVER")

	if pserver != "" && !NoPreBuildFiles {
		url, err := url.Parse(pserver)
		if err != nil {
			panic(err)
		}
		qq.Println("@using_dev_proxy", pserver)

		proxy = httputil.NewSingleHostReverseProxy(url)
		return func(ctx *gin.Context) {
			qq.Println("[PROXY]", ctx.Request.URL.String())
			proxy.ServeHTTP(ctx.Writer, ctx.Request)
		}

	}
	qq.Println("@not_using_dev_proxy")

	return func(ctx *gin.Context) {

		ppath := strings.TrimSuffix(strings.TrimPrefix(ctx.Request.URL.Path, "/zz/pages"), "/")

		if ppath == "" {
			ppath = "index.html"
		}

		pitems := strings.Split(ppath, "/")
		lastpath := pitems[len(pitems)-1]

		if !strings.Contains(lastpath, ".") {
			ppath = ppath + ".html"
		}

		qq.Println("@FILE ==>", ppath)

		out, err := output.BuildProd.ReadFile(path.Join("output/build", ppath))
		if err != nil {
			qq.Println("@open_err", err.Error())
			return
		}

		httpx.WriteFile(ppath, out, ctx)
	}
}
