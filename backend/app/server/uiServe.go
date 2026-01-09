package server

import (
	"io"
	"path"
	"strings"

	"github.com/blue-monads/potatoverse/backend/app/server/assets"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/frontend"

	"github.com/gin-gonic/gin"
)

var (
	DEV_MODE = true
)

// during dev we just proxy to dev server running otherwise serve files from build folder
func (s *Server) pages(z *gin.RouterGroup) {
	rfunc := assets.PagesRoutesServer()

	z.GET("/pages", rfunc)
	z.GET("/pages/*files", rfunc)
	z.GET("/lib/*file", func(ctx *gin.Context) {

		qq.Println("@lib/1")

		file := ctx.Param("file")

		file = strings.TrimPrefix(file, "/")

		fout, err := frontend.BuildProd.Open(path.Join("output", file))
		if err != nil {
			qq.Println("@lib/2", err.Error())
			return
		}

		qq.Println("@lib/3")

		defer fout.Close()

		io.Copy(ctx.Writer, fout)

	})

}
