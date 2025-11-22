package xtypes

import (
	"github.com/gin-gonic/gin"
)

type Capability interface {
	Handle(ctx *gin.Context)
	ListActions() ([]string, error)
	Execute(name string, params LazyData) (map[string]any, error)
	Reload(opts LazyData) (Capability, error)
	Close() error
}

type CapabilityBuilderFactory struct {
	Name         string
	Icon         string
	OptionFields []CapabilityOptionField

	Builder func(app App) (CapabilityBuilder, error)
}

type CapabilityOptionField struct {
	Name        string `json:"name"`
	Key         string `json:"key"`
	Description string `json:"description"`
	// text, number, date, api_key, boolean, select, multi_select, textarea, object
	Type     string   `json:"type"`
	Default  string   `json:"default"`
	Options  []string `json:"options"`
	Required bool     `json:"required"`
}

type CapabilityBuilder interface {
	Name() string
	Build(installId, spaceId int64, opts LazyData) (Capability, error)
	Serve(ctx *gin.Context)
}

type CapabilityHub interface {
	List(spaceId int64) ([]string, error)
	Execute(installId, spaceId int64, gname, method string, params LazyData) (map[string]any, error)
	Methods(installId, spaceId int64, gname string) ([]string, error)
}
