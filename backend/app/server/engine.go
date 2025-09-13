package server

import "github.com/gin-gonic/gin"

type Engine interface {
	ServeSpaceFile(ctx *gin.Context)
	ServePluginFile(ctx *gin.Context)
	SpaceApi(ctx *gin.Context)
	PluginApi(ctx *gin.Context)
}

func (a *Server) handleSpaceFile(ctx *gin.Context) {

}

func (a *Server) handlePluginFile(ctx *gin.Context) {}

func (a *Server) handleSpaceApi(ctx *gin.Context) {}

func (a *Server) handlePluginApi(ctx *gin.Context) {}
