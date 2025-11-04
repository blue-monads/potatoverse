package xtypes

import (
	"io/fs"
	"log/slog"

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

type BuilderOption struct {
	App App

	Logger *slog.Logger
}

type Builder func(opt BuilderOption) (*Defination, error)

type Defination struct {
	Name            string
	Slug            string
	Info            string
	Icon            string
	Version         string
	AssetData       fs.FS
	AssetDataPrefix string

	LinkPattern string

	OnInit       func(sid int64) error
	IsInitilized func(sid int64) (bool, error)
	OnDeInit     func(sid int64) error
	OnClose      func() error
}
