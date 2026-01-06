package xtypes

import (
	"log/slog"
	"os"

	"github.com/blue-monads/turnix/backend/xtypes/lazydata"
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

type HttpEventOptions struct {
	SpaceId     int64
	EventType   string // http
	HandlerName string
	Params      map[string]string
	Request     *gin.Context
}

type ActionEventOptions struct {
	SpaceId    int64
	EventType  string // ws, ws_callback, event_target, mcp_call
	ActionName string
	Params     map[string]string
	Request    ActionRequest
}

type Engine interface {
	GetCapabilityHub() any
	GetBuddyHub() any
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

	EmitHttpEvent(opts *HttpEventOptions) error
	EmitActionEvent(opts *ActionEventOptions) error
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

// HttpEvent handled by on_http method
type HttpEvent struct {
	EventType   string // http, api
	HandlerName string
	Params      map[string]string
	Request     *gin.Context
}

// ActionEvent handled by on_action
type ActionEvent struct {
	EventType  string // capability, event_target
	ActionName string
	Params     map[string]string
	Request    ActionRequest
}

type ActionRequest interface {
	ListActions() ([]string, error)
	ExecuteAction(name string, params lazydata.LazyData) (any, error)
}

type Executor interface {
	Cleanup()
	GetDebugData() map[string]any

	HandleHttp(event *HttpEvent) error
	HandleAction(event *ActionEvent) error
}

// RootExecutor types

type RootExecutor interface {
	ServeRoot(ctx *gin.Context)
}

type RootExecutorFactory func(app App) (RootExecutor, error)
