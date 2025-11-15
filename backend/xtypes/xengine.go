package xtypes

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

type Engine interface {
	GetCapabilityHub() any
	GetDebugData() map[string]any
	LoadRoutingIndex()

	PluginApi(ctx *gin.Context)
	ServePluginFile(ctx *gin.Context)

	ServeCapability(ctx *gin.Context)
	ServeCapabilityRoot(ctx *gin.Context)

	ServeSpaceFile(ctx *gin.Context)
	SpaceApi(ctx *gin.Context)

	PublishEvent(installId int64, name string, payload []byte) error
}

type ExecutorBuilderOption struct {
	App App

	Logger *slog.Logger

	WorkingFolder    string
	SpaceId          int64
	PackageVersionId int64
	InstalledId      int64
	FsRoot           *os.Root
}

type ExecutorBuilder struct {
	Name  string
	Icon  string
	Build func(opt ExecutorBuilderOption) (*Executor, error)
}

type HttpExecution struct {
	HandlerName string
	Params      map[string]string
	Request     *gin.Context
}

type GenericExecution struct {
	Type       string // ws, ws_callback, event_target, mcp_call
	ActionName string
	Params     map[string]string
	Context    GenericContext
}

type GenericContext interface {
	ListActions() ([]string, error)
	ExecuteAction(name string, params LazyData) (map[string]any, error)
}

type Executor interface {
	HandleHttp(event HttpExecution) error
	HandleGeneric(event GenericExecution) error
}
