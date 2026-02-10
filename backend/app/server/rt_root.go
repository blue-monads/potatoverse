package server

import "github.com/gin-gonic/gin"

func (s *Server) RootRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Redirect(302, "/zz/pages")
	}
}
