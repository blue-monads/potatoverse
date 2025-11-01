package xtypes

import "github.com/gin-gonic/gin"

type Capability interface {
	Name() string
	Handle(ctx *gin.Context)
	ListActions() ([]string, error)
	GetActionMeta(name string) (map[string]any, error)
	ExecuteAction(name string, params LazyData) (map[string]any, error)
}

type CapabilityBuilderFactory func(app App) (CapabilityBuilder, error)

type CapabilityBuilder interface {
	Build(spaceId int64) (Capability, error)
	Serve(ctx *gin.Context)
}

type CapabilityHub interface {
	List(spaceId int64) ([]string, error)
	GetMeta(spaceId int64, gname, method string) (map[string]any, error)
	Execute(spaceId int64, gname, method string, params LazyData) (map[string]any, error)
	Methods(spaceId int64, gname string) ([]string, error)
}
