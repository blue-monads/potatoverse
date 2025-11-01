package xtypes

import (
	"github.com/gin-gonic/gin"
)

type Engine interface {
	GetAddonHub() any
	GetDebugData() map[string]any
	LoadRoutingIndex() error

	PluginApi(ctx *gin.Context)
	ServePluginFile(ctx *gin.Context)

	ServeAddon(ctx *gin.Context)
	ServeAddonRoot(ctx *gin.Context)

	ServeSpaceFile(ctx *gin.Context)
	SpaceApi(ctx *gin.Context)
}

// add on

type AddOnHub interface {
	List(spaceId int64) ([]string, error)
	GetMeta(spaceId int64, gname, method string) (map[string]any, error)
	Execute(spaceId int64, gname, method string, params LazyData) (map[string]any, error)
	Methods(spaceId int64, gname string) ([]string, error)
}

type LazyData interface {
	AsMap() (map[string]any, error)
	// AsJSON struct target
	AsJson(target any) error
}

type AddOn interface {
	Name() string
	Handle(ctx *gin.Context)
	List() ([]string, error)
	GetMeta(name string) (map[string]any, error)
	Execute(method string, params LazyData) (map[string]any, error)
}

type AddOnBuilderFactory func(app App) (AddOnBuilder, error)

type AddOnBuilder interface {
	Build(spaceId int64) (AddOn, error)
	Serve(ctx *gin.Context)
}

type ExecutionContextAction interface {
	Methods() ([]string, error)
	GetMeta(name string) (map[string]any, error)
	Execute(method string, params LazyData) (map[string]any, error)
}
