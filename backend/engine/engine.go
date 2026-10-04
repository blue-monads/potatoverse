package engine

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/caphub"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/remotehub"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/sighub"
	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	xutils "github.com/blue-monads/potatoverse/backend/utils"
	"github.com/blue-monads/potatoverse/backend/utils/libx/httpx"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
)

var _ xtypes.Engine = (*Engine)(nil)

type Engine struct {
	db            datahub.Database
	workingFolder string

	runtime     Runtime
	spaceRouter *SpaceRouter
	rootRouter  *RootRouter

	logger *slog.Logger

	app xtypes.App

	repoHub *repohub.RepoHub

	sigHub *sighub.SigHub

	capHub *caphub.CapabilityHub

	remoteHub *remotehub.RemoteHub

	HttpPort int
}

type EngineOption struct {
	DB            datahub.Database
	WorkingFolder string
	Logger        *slog.Logger
	Repos         []xtypes.RepoOptions
	HttpPort      int
}

func NewEngine(opt EngineOption) *Engine {

	elogger := opt.Logger.With("module", "engine")

	e := &Engine{
		db:            opt.DB,
		workingFolder: opt.WorkingFolder,
		runtime: Runtime{
			activeExecs:     make(map[int64]*RunningExec),
			activeExecsLock: sync.RWMutex{},
			builders:        make(map[string]xtypes.ExecutorBuilder),
		},
		logger:    elogger,
		capHub:    caphub.NewCapabilityHub(),
		remoteHub: remotehub.NewRemoteHub(),
		HttpPort:  opt.HttpPort,

		repoHub: repohub.NewRepoHub(opt.Repos, elogger.With("service", "repo_hub"), opt.HttpPort),
	}

	e.runtime.parent = e
	e.spaceRouter = NewSpaceRouter(e)
	e.rootRouter = NewRootRouter(e)

	return e
}

func (e *Engine) GetDebugData() map[string]any {
	var routingIndexCopy map[string]*SpaceRouteIndexItem
	if e.spaceRouter != nil {
		routingIndexCopy = e.spaceRouter.GetRoutingIndexCopy()
	}

	return map[string]any{
		"runtime_data":  e.runtime.GetDebugData(),
		"routing_index": routingIndexCopy,
	}
}

func (e *Engine) EmitHttpEvent(opts *xtypes.HttpEventOptions) error {
	return e.runtime.ExecHttp(opts)
}

func (e *Engine) EmitActionEvent(opts *xtypes.ActionEventOptions) error {
	return e.runtime.ExecAction(opts)
}

func (e *Engine) Start(app xtypes.App) error {
	e.app = app
	e.runtime.parent = e
	e.logger = app.Logger().With("module", "engine")

	bfactories := registry.GetExecutorBuilderFactories()

	for name, factory := range bfactories {
		builder, err := factory(app)
		if err != nil {
			return err
		}
		e.runtime.builders[name] = builder
	}

	// Initialize capabilities hub
	err := e.capHub.Init(app)
	if err != nil {
		return err
	}

	e.sigHub = sighub.NewSigHub(app)
	err = e.sigHub.Start()
	if err != nil {
		return err
	}

	err = e.repoHub.Run(app)
	if err != nil {
		return err
	}

	e.remoteHub.Init(app)

	if e.spaceRouter != nil {
		e.spaceRouter.Start()
	}

	if e.rootRouter != nil {
		_ = e.rootRouter.ReloadIndex()
	}

	time.Sleep(2 * time.Second)

	return nil
}

func (e *Engine) Close() {
	if e.spaceRouter != nil {
		e.spaceRouter.Close()
	}
	if e.sigHub != nil {
		e.sigHub.Stop()
	}
	if e.capHub != nil {
		e.capHub.Close()
	}
	e.runtime.CloseAll()
}

func (e *Engine) ServeRootSpace(ctx *gin.Context) {
	e.rootRouter.Serve(ctx)
}

func (e *Engine) ServeSpaceFile(ctx *gin.Context) {
	e.spaceRouter.ServeSpaceFile(ctx, 0)
}

func (e *Engine) GetSpaceRouter() *SpaceRouter {
	return e.spaceRouter
}

func (e *Engine) GetRootRouter() *RootRouter {
	return e.rootRouter
}

func (e *Engine) ReloadRootRoutingIndex() error {
	if e.rootRouter != nil {
		return e.rootRouter.ReloadIndex()
	}
	return nil
}

func (e *Engine) LoadRoutingIndex() {
	if e.spaceRouter != nil {
		e.spaceRouter.LoadRoutingIndex()
	}
}

func (e *Engine) LoadRoutingIndexForPackages(installedId int64) {
	if e.spaceRouter != nil {
		e.spaceRouter.LoadRoutingIndexForPackages(installedId)
	}
}

func (e *Engine) GetPluginLoaderScript(spaceKey string) string {
	if e.spaceRouter != nil {
		return e.spaceRouter.GetPluginLoaderScript(spaceKey)
	}
	return ""
}

func (e *Engine) getIndex(spaceKey string, spaceId int64) *SpaceRouteIndexItem {
	if e.spaceRouter != nil {
		return e.spaceRouter.getIndex(spaceKey, spaceId)
	}
	return nil
}

func (e *Engine) getIndexRetry(spaceKey string, spaceId int64) *SpaceRouteIndexItem {
	if e.spaceRouter != nil {
		return e.spaceRouter.getIndexRetry(spaceKey, spaceId)
	}
	return nil
}

func (e *Engine) ServePluginFile(ctx *gin.Context) {

}

func (e *Engine) ServeCapability(ctx *gin.Context) {
	spaceKey := ctx.Param("space_key")
	capabilityName := ctx.Param("capability_name")

	spaceId := xutils.ExtractSpaceId(ctx.Request.Host)

	index := e.getIndex(spaceKey, spaceId)

	if index == nil {
		httpx.WriteErr(ctx, errors.New("space not found"))
		return
	}

	e.capHub.Handle(index.installedId, spaceId, capabilityName, ctx)

}

func (e *Engine) ServeCapabilityRoot(ctx *gin.Context) {
	capabilityName := ctx.Param("capability_name")
	e.capHub.HandleRoot(capabilityName, ctx)
}

func (e *Engine) SpaceApi(ctx *gin.Context) {
	spaceKey := ctx.Param("space_key")
	spaceId := xutils.ExtractSpaceId(ctx.Request.Host)

	qq.Println("@SpaceApi/3", spaceKey, spaceId)

	sIndex := e.getIndex(spaceKey, spaceId)

	if sIndex == nil {
		httpx.WriteErrString(ctx, "space not found")
		return
	}

	err := e.runtime.ExecHttpQ(
		sIndex.installedId,
		sIndex.packageVersionId,
		sIndex.spaceId,
		ctx,
	)
	if err != nil {
		e.logger.Error("error executing http request", "error", err)
		ctx.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
}

func (e *Engine) PluginApi(ctx *gin.Context) {
	spaceKey := ctx.Param("space_key")
	pluginKey := ctx.Param("plugin_key")
	spaceId := xutils.ExtractSpaceId(ctx.Request.Host)

	qq.Println("@PluginApi/1", spaceKey, pluginKey, spaceId)

	sIndex := e.getIndex(spaceKey, spaceId)
	if sIndex == nil {
		httpx.WriteErrString(ctx, "host space not found")
		return
	}

	plugins, err := e.db.GetSpaceOps().ListSpacePlugins(sIndex.installedId, sIndex.spaceId)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	var targetPluginSpace *dbmodels.Space
	var matchedPluginId int64
	for _, plug := range plugins {
		pspace, err := e.db.GetSpaceOps().GetSpace(plug.TargetSpaceID)
		if err != nil || pspace == nil {
			continue
		}
		if pspace.NamespaceKey == pluginKey {
			targetPluginSpace = pspace
			matchedPluginId = plug.ID
			break
		}
	}

	if targetPluginSpace == nil {
		httpx.WriteErrString(ctx, "plugin not found")
		return
	}

	pluginPkg, err := e.db.GetPackageInstallOps().GetPackage(targetPluginSpace.InstalledId)
	if err != nil || pluginPkg == nil {
		httpx.WriteErrString(ctx, "plugin package not found")
		return
	}

	reqId, _ := xutils.GenerateRandomString(12)
	remoteCtxToken, err := e.remoteHub.SignRemoteCtxToken(&remotehub.RemoteCtxClaim{
		TargetSpaceId:          sIndex.spaceId,
		TargetPackageId:        sIndex.installedId,
		TargetPackageVersionId: sIndex.packageVersionId,
		PluginSpaceId:          targetPluginSpace.ID,
		PluginId:               matchedPluginId,
		RequestID:              reqId,
	})
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	params := map[string]string{
		"remote_ctx_token": remoteCtxToken,
	}

	e.runtime.ExecHttpWithParams(
		pluginPkg.ID,
		pluginPkg.ActiveInstallID,
		targetPluginSpace.ID,
		params,
		ctx,
	)
}

type SpaceInfo struct {
	ID            int64  `json:"id"`
	NamespaceKey  string `json:"namespace_key"`
	OwnsNamespace bool   `json:"owns_namespace"`
	PackageName   string `json:"package_name"`
}

func (e *Engine) SpaceInfo(nsKey string, hostName string) (*SpaceInfo, error) {

	qq.Println("@SpaceInfo/1", nsKey, hostName)

	var index *SpaceRouteIndexItem

	if hostName != "" {
		spaceId := xutils.ExtractSpaceId(hostName)

		// fixme => should i check if spaceId == 0 ?

		index = e.getIndex(nsKey, spaceId)

	}

	if index == nil {
		return nil, errors.New("space not found")
	}

	space, err := e.db.GetSpaceOps().GetSpace(index.spaceId)
	if err != nil {
		return nil, err
	}

	pkg, err := e.db.GetPackageInstallOps().GetPackage(space.InstalledId)
	if err != nil {
		return nil, err
	}

	return &SpaceInfo{
		ID:            space.ID,
		NamespaceKey:  space.NamespaceKey,
		OwnsNamespace: true,
		PackageName:   pkg.Name,
	}, nil

}

func (e *Engine) GetCapabilityDefinitions() []caphub.CapabilityDefination {
	return e.capHub.Definations()
}

func (e *Engine) GetCapabilityHub() any {
	return e.capHub
}

func (e *Engine) GetRepoHub() *repohub.RepoHub {
	return e.repoHub
}

func (e *Engine) GetRemoteHub() *remotehub.RemoteHub {
	return e.remoteHub
}

func (e *Engine) PublishEvent(opts *xtypes.EventOptions) error {
	return e.PublishSignal(&xtypes.SignalOptions{
		SignalKey:        opts.Name,
		EmitterInstallId: opts.InstallId,
		EmitterSpaceId:   opts.SpaceId,
		Payload:          opts.Payload,
	})
}

func (e *Engine) RefreshEventIndex() {
	e.RefreshSignalIndex()
}

func (e *Engine) PublishSignal(opts *xtypes.SignalOptions) error {
	if e.sigHub == nil {
		return nil
	}
	return e.sigHub.Publish(opts)
}

func (e *Engine) RefreshSignalIndex() {
	if e.sigHub != nil {
		e.sigHub.RefreshFullIndex()
	}
}
