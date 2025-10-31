package addons

import (
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type Hub interface {
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

type BuilderFactory func(app xtypes.App) (Builder, error)

type Builder interface {
	Build(spaceId int64) (AddOn, error)
	Serve(ctx *gin.Context)
}
