package xtypes

import (
	"github.com/gin-gonic/gin"
)

type Engine interface {
	GetCapabilityHub() any
	GetDebugData() map[string]any
	LoadRoutingIndex() error

	PluginApi(ctx *gin.Context)
	ServePluginFile(ctx *gin.Context)

	ServeCapability(ctx *gin.Context)
	ServeCapabilityRoot(ctx *gin.Context)

	ServeSpaceFile(ctx *gin.Context)
	SpaceApi(ctx *gin.Context)
}

type LazyData interface {
	AsMap() (map[string]any, error)
	// AsJSON struct target
	AsJson(target any) error
}

// capability
