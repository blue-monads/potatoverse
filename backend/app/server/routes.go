package server

import (
	"embed"
	_ "embed"
	"net/http"
	"path"

	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/docs"
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

	a.buddyRoutes.AttachRoutes(root)

	coreApi.GET("/global.js", a.getGlobalJS)

	root.GET("/static/*files", func(c *gin.Context) {
		filePath := c.Param("files")
		if len(filePath) > 0 && filePath[0] == '/' {
			filePath = filePath[1:]
		}

		fullPath := "static/" + filePath

		c.FileFromFS(fullPath, http.FS(StaticFiles))
	})

	coreApi.GET("/docs/*path", func(c *gin.Context) {
		ppath := c.Param("path")
		fpath := path.Join("contents", ppath)

		qq.Println("@fpath", fpath)
		qq.Println("@ppath", ppath)

		c.FileFromFS(fpath, http.FS(docs.Docs))
	})

}

func (a *Server) authRoutes(g *gin.RouterGroup) {

	g.POST("/login", a.login)
	g.GET("/invite/:token", a.getInviteInfo)
	g.POST("/invite/:token", a.acceptInvite)

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

	// User Groups
	g.GET("/groups", a.withAccessTokenFn(a.listUserGroups))
	g.GET("/groups/:name", a.withAccessTokenFn(a.getUserGroup))
	g.POST("/groups", a.withAccessTokenFn(a.addUserGroup))
	g.PUT("/groups/:name", a.withAccessTokenFn(a.updateUserGroup))
	g.DELETE("/groups/:name", a.withAccessTokenFn(a.deleteUserGroup))

	g.GET("/messages", a.withAccessTokenFn(a.listUserMessages))
	g.GET("/messages/new", a.withAccessTokenFn(a.queryNewMessages))
	g.GET("/messages/history", a.withAccessTokenFn(a.queryMessageHistory))
	g.POST("/messages/read-all", a.withAccessTokenFn(a.setAllMessagesAsRead))
	g.POST("/messages", a.withAccessTokenFn(a.sendUserMessage))
	g.GET("/messages/:id", a.withAccessTokenFn(a.getUserMessage))
	g.PUT("/messages/:id", a.withAccessTokenFn(a.updateUserMessage))
	g.DELETE("/messages/:id", a.withAccessTokenFn(a.deleteUserMessage))
	g.POST("/messages/:id/read", a.withAccessTokenFn(a.setMessageAsRead))

}

func (a *Server) selfUserRoutes(g *gin.RouterGroup) {
	g.GET("/portalData/:portal_type", a.withAccessTokenFn(a.selfUserPortalData))
	g.GET("/info", a.withAccessTokenFn(a.selfInfo))
	g.PUT("/bio", a.withAccessTokenFn(a.updateSelfBio))
}

func (a *Server) extraRoutes(g *gin.RouterGroup) {
	g.GET("/profileImage/:id/:name", a.userSvgProfileIcon)
	g.GET("/profileImage/:id", a.userSvgProfileIconById)
	g.GET("/api/gradients", a.ListGradients)
}

func (a *Server) engineRoutes(zg *gin.RouterGroup, coreApi *gin.RouterGroup) {

	spaceFile := a.handleSpaceFile()
	pluginFile := a.handlePluginFile()

	coreApi.POST("/package/install", a.withAccessTokenFn(a.InstallPackage))
	coreApi.POST("/package/install/zip", a.withAccessTokenFn(a.InstallPackageZip))
	coreApi.POST("/package/install/embed", a.withAccessTokenFn(a.InstallPackageEmbed))
	coreApi.DELETE("/package/:id", a.withAccessTokenFn(a.DeletePackage))
	coreApi.POST("/package/:id/dev-token", a.withAccessTokenFn(a.GeneratePackageDevToken))
	coreApi.POST("/package/push", a.PushPackage)
	coreApi.GET("/package/list", a.withAccessTokenFn(a.ListEPackages))
	coreApi.GET("/repo/list", a.withAccessTokenFn(a.ListRepos))
	coreApi.GET("/space/installed", a.withAccessTokenFn(a.ListInstalledSpaces))
	coreApi.POST("/space/authorize/:space_key", a.withAccessTokenFn(a.AuthorizeSpace))
	coreApi.GET("/package/:id/info", a.withAccessTokenFn(a.GetInstalledPackageInfo))

	// Package Version Files API
	coreApi.GET("/vpackage/:id/files", a.withAccessTokenFn(a.ListPackageFiles))
	coreApi.GET("/vpackage/:id/files/:fileId", a.withAccessTokenFn(a.GetPackageFile))
	coreApi.GET("/vpackage/:id/files/:fileId/download", a.withAccessTokenFn(a.DownloadPackageFile))
	coreApi.DELETE("vpackage/:id/files/:fileId", a.withAccessTokenFn(a.DeletePackageFile))
	coreApi.POST("/vpackage/:id/files/upload", a.withAccessTokenFn(a.UploadPackageFile))

	// Space KV API
	coreApi.GET("/space/:install_id/kv", a.withAccessTokenFn(a.ListSpaceKV))
	coreApi.GET("/space/:install_id/kv/:kvId", a.withAccessTokenFn(a.GetSpaceKV))
	coreApi.POST("/space/:install_id/kv", a.withAccessTokenFn(a.CreateSpaceKV))
	coreApi.PUT("/space/:install_id/kv/:kvId", a.withAccessTokenFn(a.UpdateSpaceKV))
	coreApi.DELETE("/space/:install_id/kv/:kvId", a.withAccessTokenFn(a.DeleteSpaceKV))

	// Space Files API
	coreApi.GET("/space/:install_id/files", a.withAccessTokenFn(a.ListSpaceFiles))
	coreApi.GET("/space/:install_id/files/:fileId", a.withAccessTokenFn(a.GetSpaceFile))
	coreApi.GET("/space/:install_id/files/:fileId/download", a.withAccessTokenFn(a.DownloadSpaceFile))
	coreApi.DELETE("/space/:install_id/files/:fileId", a.withAccessTokenFn(a.DeleteSpaceFile))
	coreApi.POST("/space/:install_id/files/upload", a.withAccessTokenFn(a.UploadSpaceFile))
	coreApi.POST("/space/:install_id/files/folder", a.withAccessTokenFn(a.CreateSpaceFolder))
	coreApi.POST("/space/:install_id/files/presigned", a.withAccessTokenFn(a.CreatePresignedUploadURL))

	// Space Capabilities API
	coreApi.GET("/space/:install_id/capabilities", a.withAccessTokenFn(a.ListSpaceCapabilities))
	coreApi.GET("/space/:install_id/capabilities/:capabilityId", a.withAccessTokenFn(a.GetSpaceCapability))
	coreApi.POST("/space/:install_id/capabilities", a.withAccessTokenFn(a.CreateSpaceCapability))
	coreApi.PUT("/space/:install_id/capabilities/:capabilityId", a.withAccessTokenFn(a.UpdateSpaceCapability))
	coreApi.DELETE("/space/:install_id/capabilities/:capabilityId", a.withAccessTokenFn(a.DeleteSpaceCapability))

	// Space Users API
	coreApi.GET("/space/:install_id/users", a.withAccessTokenFn(a.ListSpaceUsers))
	coreApi.GET("/space/:install_id/users/:spaceUserId", a.withAccessTokenFn(a.GetSpaceUser))
	coreApi.POST("/space/:install_id/users", a.withAccessTokenFn(a.CreateSpaceUser))
	coreApi.PUT("/space/:install_id/users/:spaceUserId", a.withAccessTokenFn(a.UpdateSpaceUser))
	coreApi.DELETE("/space/:install_id/users/:spaceUserId", a.withAccessTokenFn(a.DeleteSpaceUser))

	// Event Subscriptions API
	coreApi.GET("/space/:install_id/events", a.withAccessTokenFn(a.ListEventSubscriptions))
	coreApi.GET("/space/:install_id/events/:subscriptionId", a.withAccessTokenFn(a.GetEventSubscription))
	coreApi.POST("/space/:install_id/events", a.withAccessTokenFn(a.CreateEventSubscription))
	coreApi.PUT("/space/:install_id/events/:subscriptionId", a.withAccessTokenFn(a.UpdateEventSubscription))
	coreApi.DELETE("/space/:install_id/events/:subscriptionId", a.withAccessTokenFn(a.DeleteEventSubscription))
	coreApi.GET("/space/:install_id/spec.json", a.withAccessTokenFn(a.GetSpaceSpec))

	// Capability Types API
	coreApi.GET("/capability/types", a.withAccessTokenFn(a.ListCapabilityTypes))

	zg.POST("/file/upload-presigned", a.UploadFileWithPresigned)

	coreApi.GET("/engine/debug", a.handleEngineDebugData)
	coreApi.GET("/engine/space_info/:space_key", a.handleSpaceInfo)
	coreApi.GET("/engine/derivehost/:nskey", a.handleDeriveHost)

	zg.Any("/space/:space_key/*subpath", spaceFile)
	zg.Any("/plugin/:space_key/:plugin_id/*subpath", pluginFile)
	zg.Any("/capabilities/:space_key/:capability_name/*subpath", a.handleCapabilities)
	zg.Any("/capabilities/:space_key/:capability_name", a.handleCapabilities)
	zg.Any("/capability-root/:capability_name/*subpath", a.handleCapabilitiesRoot)

	zg.Any("/api/space/:space_key", a.handleSpaceApi)
	zg.Any("/api/space/:space_key/*subpath", a.handleSpaceApi)
	zg.Any("/api/plugin/:space_key/:plugin_id", a.handlePluginApi)
	zg.Any("/api/plugin/:space_key/:plugin_id/*subpath", a.handlePluginApi)

	zg.Any("/api/capabilities/:space_key/:capability_name", a.handleCapabilities)
	zg.Any("/api/capabilities/:space_key/:capability_name/*subpath", a.handleCapabilities)
	zg.GET("/api/capabilities/debug/:capability_name", a.withAccessTokenFn(a.handleCapabilitiesDebug))

}
