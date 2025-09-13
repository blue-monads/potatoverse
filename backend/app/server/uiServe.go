package server

import (
	"io"
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

// during dev we just proxy to dev server running otherwise serve files from build folder
func (s *Server) pages(z *gin.RouterGroup) {
	rfunc := assets.PagesRoutesServer()

	z.GET("/pages", rfunc)
	z.GET("/pages/*files", rfunc)
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
