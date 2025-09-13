package server

import (
	"github.com/gin-gonic/gin"
)

func (a *Server) bindRoutes() {

	root := a.router.Group("/z")

	root.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	coreApi := root.Group("/api/core")

	a.pages(root)
	a.extraRoutes(root)

	a.userRoutes(coreApi.Group("/user"))
	a.authRoutes(coreApi.Group("/auth"))
	a.selfUserRoutes(coreApi.Group("/self"))
	a.engineRoutes(root, coreApi)

}

func (a *Server) authRoutes(g *gin.RouterGroup) {

	g.POST("/login", a.login)

}

func (a *Server) userRoutes(g *gin.RouterGroup) {
	g.GET("/", a.withAccessTokenFn(a.listUsers))
	g.GET("/:id", a.withAccessTokenFn(a.getUser))

}

func (a *Server) selfUserRoutes(g *gin.RouterGroup) {
	g.GET("/portalData/:portal_type", a.withAccessTokenFn(a.selfUserPortalData))
}

func (a *Server) extraRoutes(g *gin.RouterGroup) {
	g.GET("/profileImage/:id/:name", a.userSvgProfileIcon)
	g.GET("/profileImage/:id", a.userSvgProfileIconById)
}

func (a *Server) engineRoutes(zg *gin.RouterGroup, coreApi *gin.RouterGroup) {

	spaceFile := a.handleSpaceFile()
	pluginFile := a.handlePluginFile()

	zg.GET("/space/:space_key/*files", spaceFile)

	zg.GET("/plugin/:space_key/:plugin_id/*files", pluginFile)

	zg.GET("/api/space/:space_key", a.handleSpaceApi)
	zg.GET("/api/plugin/:space_key/:plugin_id", a.handlePluginApi)

	// internal file serve

	zg.GET("/pages/space/:space_key")

	coreApi.POST("/package/install", a.InstallPackage)

}
