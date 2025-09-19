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

	coreApi.POST("/package/install", a.withAccessTokenFn(a.InstallPackage))
	coreApi.POST("/package/install/zip", a.withAccessTokenFn(a.InstallPackageZip))
	coreApi.POST("/package/install/embed", a.withAccessTokenFn(a.InstallPackageEmbed))
	coreApi.DELETE("/package/:id", a.withAccessTokenFn(a.DeletePackage))
	coreApi.GET("/package/list", a.withAccessTokenFn(a.ListEPackages))
	coreApi.GET("/space/installed", a.withAccessTokenFn(a.ListInstalledSpaces))

	coreApi.GET("/engine/debug", a.handleEngineDebugData)
	coreApi.GET("/engine/space_info/:space_key", a.handleSpaceInfo)

	zg.GET("/space/:space_key/*files", spaceFile)

	zg.GET("/plugin/:space_key/:plugin_id/*files", pluginFile)

	zg.GET("/api/space/:space_key", a.handleSpaceApi)
	zg.GET("/api/space/:space_key/*subpath", a.handleSpaceApi)
	zg.GET("/api/plugin/:space_key/:plugin_id", a.handlePluginApi)
	zg.GET("/api/plugin/:space_key/:plugin_id/*subpath", a.handlePluginApi)

}
