package server

import (
	"github.com/gin-gonic/gin"
)

func (a *Server) bindRoutes() {

	root := a.router.Group("/z")

	root.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	a.authRoutes(root)
	a.pages(root)

}

func (a *Server) authRoutes(g *gin.RouterGroup) {

}
