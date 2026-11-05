package server

import (
	"strconv"

	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/gin-gonic/gin"
)

// ListSpacePlugins lists all plugins attached to a space/package
func (a *Server) ListSpacePlugins(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	spaceIdParam := ctx.Query("space_id")
	var spaceId int64
	if spaceIdParam != "" {
		spaceId, _ = strconv.ParseInt(spaceIdParam, 10, 64)
	}

	return a.ctrl.ListSpacePlugins(installId, spaceId)
}

// ListAvailablePlugins lists all AppPlugin spaces available to be plugged
func (a *Server) ListAvailablePlugins(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	spaceIdParam := ctx.Query("space_id")
	var spaceId int64
	if spaceIdParam != "" {
		spaceId, _ = strconv.ParseInt(spaceIdParam, 10, 64)
	}

	return a.ctrl.ListAvailableAppPlugins(installId, spaceId)
}

// GetSpacePlugin gets a specific space plugin by ID
func (a *Server) GetSpacePlugin(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	pluginId, err := strconv.ParseInt(ctx.Param("pluginId"), 10, 64)
	if err != nil {
		return nil, err
	}

	return a.ctrl.GetSpacePluginByID(installId, pluginId)
}

// CreateSpacePlugin attaches an AppPlugin to a space
func (a *Server) CreateSpacePlugin(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	spaceIdParam := ctx.Query("space_id")
	var spaceId int64
	if spaceIdParam != "" {
		spaceId, _ = strconv.ParseInt(spaceIdParam, 10, 64)
	}

	var data map[string]any
	if err := ctx.ShouldBindJSON(&data); err != nil {
		return nil, err
	}

	return a.ctrl.CreateSpacePlugin(installId, spaceId, data)
}

// UpdateSpacePlugin updates an existing space plugin connection
func (a *Server) UpdateSpacePlugin(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	pluginId, err := strconv.ParseInt(ctx.Param("pluginId"), 10, 64)
	if err != nil {
		return nil, err
	}

	var data map[string]any
	if err := ctx.ShouldBindJSON(&data); err != nil {
		return nil, err
	}

	return a.ctrl.UpdateSpacePluginByID(installId, pluginId, data)
}

// DeleteSpacePlugin unplugs a plugin from a space
func (a *Server) DeleteSpacePlugin(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	pluginId, err := strconv.ParseInt(ctx.Param("pluginId"), 10, 64)
	if err != nil {
		return nil, err
	}

	err = a.ctrl.DeleteSpacePluginByID(installId, pluginId)
	if err != nil {
		return nil, err
	}

	return map[string]string{"status": "deleted"}, nil
}

// ServePluginLoaders returns the concatenated plugin loader script for a space
func (a *Server) ServePluginLoaders(ctx *gin.Context) {
	spaceKey := ctx.Param("space_key")
	if spaceKey == "" {
		ctx.Data(400, "application/javascript", []byte("// space key is required\n"))
		return
	}

	var script string
	if a.engine != nil {
		script = a.engine.GetPluginLoaderScript(spaceKey)
	}

	if script == "" {
		ctx.Data(200, "application/javascript", []byte("// no plugins loaded\n"))
		return
	}
	ctx.Data(200, "application/javascript", []byte(script))
}
