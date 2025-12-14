package xtypes

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

type EventOptions struct {
	InstallId   int64
	SpaceId     int64
	Name        string
	Payload     []byte
	ResourceId  string
	CollapseKey string
}

// Engine types

type EngineHttpExecution struct {
	SpaceId     int64
	HandlerName string
	Params      map[string]string
	Request     *gin.Context
}

type EngineActionExecution struct {
	SpaceId    int64
	ActionType string // ws, ws_callback, event_target, mcp_call
	ActionName string
	Params     map[string]string
	Request    ActionRequest
}

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

	PublishEvent(opts *EventOptions) error
	RefreshEventIndex()

	ExecHttp(opts *EngineHttpExecution) error
	ExecAction(opts *EngineActionExecution) error
}

// Executor types

type ExecutorBuilderOption struct {
	Logger *slog.Logger

	WorkingFolder    string
	SpaceId          int64
	PackageVersionId int64
	InstalledId      int64
	FsRoot           *os.Root
}

type ExecutorBuilderFactory func(app App) (ExecutorBuilder, error)

type ExecutorBuilder interface {
	Name() string
	Icon() string
	Build(opt *ExecutorBuilderOption) (Executor, error)
}

type HttpExecution struct {
	HandlerName string
	Params      map[string]string
	Request     *gin.Context
}

type ActionExecution struct {
	ActionType string // ws, ws_callback, event_target, mcp_call
	ActionName string
	Params     map[string]string
	Request    ActionRequest
}

type ActionRequest interface {
	ListActions() ([]string, error)
	ExecuteAction(name string, params LazyData) (map[string]any, error)
}

type Executor interface {
	Cleanup()
	GetDebugData() map[string]any

	HandleHttp(event HttpExecution) error
	HandleAction(event ActionExecution) error
}
