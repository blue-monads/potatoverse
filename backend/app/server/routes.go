package server

import (
	"github.com/gin-gonic/gin"
)

func (a *Server) bindRoutes() {

	root := a.router.Group("/z")

	a.authRoutes(root)

}

func (a *Server) authRoutes(g *gin.RouterGroup) {

}
