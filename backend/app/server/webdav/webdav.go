package webdav

import (
	"net/http"

	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/webdav"
)

type WebdavServer struct {
	RootPath string
	Prefix   string
	Handler  *webdav.Handler
}

func New(rootPath, prefix string) *WebdavServer {
	return &WebdavServer{
		RootPath: rootPath,
		Prefix:   prefix,
	}
}

func (s *WebdavServer) Build() {
	lockSystem := webdav.NewMemLS()

	handler := &webdav.Handler{
		Prefix:     s.Prefix,
		FileSystem: webdav.Dir(s.RootPath),
		LockSystem: lockSystem,
		Logger: func(r *http.Request, err error) {
			if err != nil {
				qq.Println("webdav error:", r.Method, r.URL.Path, err)
			}
		},
	}

	s.Handler = handler

}

func (s *WebdavServer) Handle(ctx *gin.Context) {
	s.Handler.ServeHTTP(ctx.Writer, ctx.Request)
}
