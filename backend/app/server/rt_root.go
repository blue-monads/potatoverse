package server

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Server) RootRoute() gin.HandlerFunc {

	return func(c *gin.Context) {

		if strings.HasPrefix(c.Request.URL.Path, "/zz/") {
			c.Redirect(302, "/zz/pages")
		} else {
			s.engine.ServeRootSpace(c)
		}

	}
}
