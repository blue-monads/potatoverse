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
	RefreshEventIndex()
}

type ExecutorBuilderOption struct {
	Logger *slog.Logger

	WorkingFolder    string
	SpaceId          int64
	PackageVersionId int64
	InstalledId      int64
	FsRoot           *os.Root
}

type ExecutorBuilderFactory func(app App) (ExecutorBuilder, error)

type ExecutorBuilder struct {
	Name  string
	Icon  string
	Build func(opt *ExecutorBuilderOption) (Executor, error)
}

type HttpExecution struct {
	HandlerName string
	Params      map[string]string
	Request     *gin.Context
}

type EventExecution struct {
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
	Cleanup()
	GetDebugData() map[string]any

	HandleHttp(event HttpExecution) error
	HandleEvent(event EventExecution) error
}
