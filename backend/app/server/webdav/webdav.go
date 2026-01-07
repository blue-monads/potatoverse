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

	FuncVerifyAuth func(ctx *gin.Context) (bool, error)

	handler    *webdav.Handler
	lockSystem webdav.LockSystem
}

// New creates a new WebdavServer with the given root path and URL prefix.
func New(rootPath, prefix string) *WebdavServer {
	return &WebdavServer{
		RootPath: rootPath,
		Prefix:   prefix,
	}
}

// AttachRoutes attaches the WebDAV handler to the given Gin router group.
// The router group's base path should match the Prefix configured in WebdavServer.
func (s *WebdavServer) AttachRoutes(router *gin.RouterGroup) {
	s.lockSystem = webdav.NewMemLS()

	s.handler = &webdav.Handler{
		Prefix:     s.Prefix,
		FileSystem: webdav.Dir(s.RootPath),
		LockSystem: s.lockSystem,
		Logger: func(r *http.Request, err error) {
			if err != nil {
				qq.Println("webdav error:", r.Method, r.URL.Path, err)
			}
		},
	}

	router.Any("/webdav/*path", s.handleWebDAV)
}

func (s *WebdavServer) handleWebDAV(ctx *gin.Context) {

	// check authentication

	s.handler.ServeHTTP(ctx.Writer, ctx.Request)
}
