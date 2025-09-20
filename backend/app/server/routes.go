package server

import (
	"embed"
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed all:static/*
var StaticFiles embed.FS

func (a *Server) bindRoutes() {

	root := a.router.Group("/zz")

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

	root.GET("/static/*files", func(c *gin.Context) {
		filePath := c.Param("files")
		if len(filePath) > 0 && filePath[0] == '/' {
			filePath = filePath[1:]
		}

		fullPath := "static/" + filePath

		c.FileFromFS(fullPath, http.FS(StaticFiles))
	})

}

func (a *Server) authRoutes(g *gin.RouterGroup) {

	g.POST("/login", a.login)

}

func (a *Server) userRoutes(g *gin.RouterGroup) {
	g.GET("/", a.withAccessTokenFn(a.listUsers))
	g.GET("/:id", a.withAccessTokenFn(a.getUser))

	// User Invites
	g.GET("/invites", a.withAccessTokenFn(a.listUserInvites))
	g.GET("/invites/:id", a.withAccessTokenFn(a.getUserInvite))
	g.POST("/invites", a.withAccessTokenFn(a.addUserInvite))
	g.PUT("/invites/:id", a.withAccessTokenFn(a.updateUserInvite))
	g.DELETE("/invites/:id", a.withAccessTokenFn(a.deleteUserInvite))
	g.POST("/invites/:id/resend", a.withAccessTokenFn(a.resendUserInvite))

	// Create User Directly
	g.POST("/create", a.withAccessTokenFn(a.createUserDirectly))

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
	coreApi.POST("/space/authorize/:space_key", a.withAccessTokenFn(a.AuthorizeSpace))

	// Package Files API
	coreApi.GET("/package/:id/files", a.withAccessTokenFn(a.ListPackageFiles))
	coreApi.GET("/package/:id/files/:fileId", a.withAccessTokenFn(a.GetPackageFile))
	coreApi.GET("/package/:id/files/:fileId/download", a.withAccessTokenFn(a.DownloadPackageFile))
	coreApi.DELETE("/package/:id/files/:fileId", a.withAccessTokenFn(a.DeletePackageFile))
	coreApi.POST("/package/:id/files/upload", a.withAccessTokenFn(a.UploadPackageFile))

	// Space KV API
	coreApi.GET("/space/:id/kv", a.withAccessTokenFn(a.ListSpaceKV))
	coreApi.GET("/space/:id/kv/:kvId", a.withAccessTokenFn(a.GetSpaceKV))
	coreApi.POST("/space/:id/kv", a.withAccessTokenFn(a.CreateSpaceKV))
	coreApi.PUT("/space/:id/kv/:kvId", a.withAccessTokenFn(a.UpdateSpaceKV))
	coreApi.DELETE("/space/:id/kv/:kvId", a.withAccessTokenFn(a.DeleteSpaceKV))

	// Space Files API
	coreApi.GET("/space/:id/files", a.withAccessTokenFn(a.ListSpaceFiles))
	coreApi.GET("/space/:id/files/:fileId", a.withAccessTokenFn(a.GetSpaceFile))
	coreApi.GET("/space/:id/files/:fileId/download", a.withAccessTokenFn(a.DownloadSpaceFile))
	coreApi.DELETE("/space/:id/files/:fileId", a.withAccessTokenFn(a.DeleteSpaceFile))
	coreApi.POST("/space/:id/files/upload", a.withAccessTokenFn(a.UploadSpaceFile))
	coreApi.POST("/space/:id/files/folder", a.withAccessTokenFn(a.CreateSpaceFolder))

	coreApi.GET("/engine/debug", a.handleEngineDebugData)
	coreApi.GET("/engine/space_info/:space_key", a.handleSpaceInfo)

	zg.GET("/space/:space_key/*files", spaceFile)

	zg.GET("/plugin/:space_key/:plugin_id/*files", pluginFile)

	zg.GET("/api/space/:space_key", a.handleSpaceApi)
	zg.GET("/api/space/:space_key/*subpath", a.handleSpaceApi)
	zg.GET("/api/plugin/:space_key/:plugin_id", a.handlePluginApi)
	zg.GET("/api/plugin/:space_key/:plugin_id/*subpath", a.handlePluginApi)

}
