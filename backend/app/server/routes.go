package server

import (
	"github.com/gin-gonic/gin"
)

func (a *Server) bindRoutes() {

	root := a.router.Group("/z")

	root.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	apig := root.Group("/api")

	a.authRoutes(apig)
	a.pages(root)

}

func (a *Server) authRoutes(g *gin.RouterGroup) {

	g.POST("/login", a.login)

}
