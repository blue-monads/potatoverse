package engine

import (
	"errors"
	"log/slog"
	"maps"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/blue-monads/turnix/backend/engine/executors/luaz"
	"github.com/blue-monads/turnix/backend/engine/hubs/caphub"
	"github.com/blue-monads/turnix/backend/engine/hubs/eventhub"
	"github.com/blue-monads/turnix/backend/engine/hubs/repohub"
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/models"
	"github.com/gin-gonic/gin"
)

var _ xtypes.Engine = (*Engine)(nil)

type Engine struct {
	db            datahub.Database
	RoutingIndex  map[string]*SpaceRouteIndexItem
	riLock        sync.RWMutex
	workingFolder string

	runtime Runtime

	logger *slog.Logger

	app xtypes.App

	repoHub *repohub.RepoHub

	eventHub *eventhub.EventHub

	capHub *caphub.CapabilityHub

	reloadPackageIds chan int64
	fullReload       chan struct{}
}

type EngineOption struct {
	DB            datahub.Database
	WorkingFolder string
	Logger        *slog.Logger
	Repos         []xtypes.RepoOptions
}

func NewEngine(opt EngineOption) *Engine {

	elogger := opt.Logger.With("module", "engine")

	e := &Engine{
		db:            opt.DB,
		workingFolder: opt.WorkingFolder,
		RoutingIndex:  make(map[string]*SpaceRouteIndexItem),
		runtime: Runtime{
			execs:     make(map[int64]*luaz.Luaz),
			execsLock: sync.RWMutex{},
		},
		logger:           elogger,
		capHub:           caphub.NewCapabilityHub(),
		riLock:           sync.RWMutex{},
		reloadPackageIds: make(chan int64, 20),
		fullReload:       make(chan struct{}, 1),

		eventHub: eventhub.NewEventHub(opt.DB),
		repoHub:  repohub.NewRepoHub(opt.Repos, elogger.With("service", "repo_hub")),
	}

	return e
}

func (e *Engine) GetDebugData() map[string]any {
	indexCopy := make(map[string]*SpaceRouteIndexItem)
	e.riLock.RLock()
	maps.Copy(indexCopy, e.RoutingIndex)
	e.riLock.RUnlock()

	return map[string]any{
		"runtime_data":  e.runtime.GetDebugData(),
		"routing_index": indexCopy,
	}

}

func (e *Engine) Start(app xtypes.App) error {
	e.app = app
	e.runtime.parent = e
	e.logger = app.Logger().With("module", "engine")

	// Initialize repo hub
	opts := app.Config().(*xtypes.AppOptions)
	if opts != nil {
		e.repoHub = repohub.NewRepoHub(opts.Repos, e.logger)
	} else {
		// Fallback: create empty repo hub
		e.repoHub = repohub.NewRepoHub([]xtypes.RepoOptions{}, e.logger)
	}

	// Initialize capabilities hub
	err := e.capHub.Init(app)
	if err != nil {
		return err
	}

	err = e.eventHub.Start()
	if err != nil {
		return err
	}

	go e.runtime.cleanupExecs()
	go e.startEloop()

	e.LoadRoutingIndex()

	return nil
}

func (e *Engine) ServeSpaceFile(ctx *gin.Context) {

	qq.Println("@ServeSpaceFile/1")

	spaceKey := ctx.Param("space_key")
	spaceId := extractDomainSpaceId(ctx.Request.Host)

	qq.Println("@ServeSpaceFile/3")

	sIndex := e.getIndex(spaceKey, spaceId)

	if sIndex == nil {
		keys := make([]string, 0)
		for key := range e.RoutingIndex {
			keys = append(keys, key)
		}

		qq.Println("@ServeSpaceFile/4", keys)
		qq.Println("@ServeSpaceFile/4")
		httpx.WriteErrString(ctx, "space not found")
		return
	}

	qq.Println("@ServeSpaceFile/5")

	switch sIndex.routeOption.RouterType {
	case "simple", "":
		e.serveSimpleRoute(ctx, sIndex)
	case "dynamic":
		qq.Println("@ServeSpaceFile/6")
		e.serveDynamicRoute(ctx, sIndex)
	default:
		httpx.WriteErrString(ctx, "router type not supported")
		return
	}

}

func (e *Engine) ServePluginFile(ctx *gin.Context) {

}

func (e *Engine) ServeCapability(ctx *gin.Context) {
	spaceKey := ctx.Param("space_key")
	capabilityName := ctx.Param("capability_name")

	spaceId := extractDomainSpaceId(ctx.Request.Host)

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
	spaceId := extractDomainSpaceId(ctx.Request.Host)

	qq.Println("@SpaceApi/3", spaceKey, spaceId)

	sIndex := e.getIndex(spaceKey, spaceId)

	if sIndex == nil {
		httpx.WriteErrString(ctx, "space not found")
		return
	}

	e.runtime.ExecHttp(spaceKey,
		sIndex.installedId,
		sIndex.packageVersionId,
		sIndex.spaceId,
		ctx,
	)

}

func (e *Engine) PluginApi(ctx *gin.Context) {

}

type SpaceInfo struct {
	ID            int64  `json:"id"`
	NamespaceKey  string `json:"namespace_key"`
	OwnsNamespace bool   `json:"owns_namespace"`
	PackageName   string `json:"package_name"`
	PackageInfo   string `json:"package_info"`
}

func (e *Engine) SpaceInfo(nsKey string, hostName string) (*SpaceInfo, error) {

	qq.Println("@SpaceInfo/1", nsKey, hostName)

	var index *SpaceRouteIndexItem

	if hostName != "" {
		spaceId := extractDomainSpaceId(hostName)
		if spaceId != 0 {
			index = e.getIndex(nsKey, spaceId)
		}
	}

	if index == nil {
		return nil, errors.New("space not found")
	}

	space, err := e.db.GetSpaceOps().GetSpace(index.spaceId)
	if err != nil {
		return nil, err
	}

	pkg, err := e.db.GetPackageInstallOps().GetPackageVersion(space.InstalledId)
	if err != nil {
		return nil, err
	}

	return &SpaceInfo{
		ID:           space.ID,
		NamespaceKey: space.NamespaceKey,
		PackageName:  pkg.Name,
		PackageInfo:  pkg.Info,
	}, nil

}

func (e *Engine) GetCapabilityDefinitions() []caphub.CapabilityDefination {
	return e.capHub.Definations()
}

// private

// s-12.example.com
var spaceIdPattern = regexp.MustCompile(`^s-(\d+)\.`)

func extractDomainSpaceId(domain string) int64 {
	qq.Println("@extractDomainSpaceId/1", domain)

	if matches := spaceIdPattern.FindStringSubmatch(domain); matches != nil {
		sid, _ := strconv.ParseInt(matches[1], 10, 64)
		return sid
	}
	return 0
}

func buildPackageFilePath(filePath string, ropt *models.PotatoRouteOptions) (string, string) {
	nameParts := strings.Split(filePath, "/")
	name := nameParts[len(nameParts)-1]
	pathParts := nameParts[:len(nameParts)-1]
	pathParts = append(pathParts, ropt.ServeFolder)

	path := strings.Join(pathParts, "/")
	path = strings.TrimLeft(path, "/")

	if ropt.TrimPathPrefix != "" {
		path = strings.TrimPrefix(path, ropt.TrimPathPrefix)
	}

	if ropt.ForceIndexHtmlFile && name == "" {
		name = "index.html"
	}

	if ropt.ForceHtmlExtension && !strings.Contains(name, ".") {
		name = name + ".html"
	}

	qq.Println("@ropt", ropt)
	qq.Println("@name", name)
	qq.Println("@path", path)

	return name, path
}

func (e *Engine) GetCapabilityHub() any {
	return e.capHub
}

func (e *Engine) GetRepoHub() *repohub.RepoHub {
	return e.repoHub
}

func (e *Engine) PublishEvent(installId int64, name string, payload []byte) error {

	return e.eventHub.Publish(installId, name, payload)
}

func (e *Engine) RefreshEventIndex() {
	e.eventHub.RefreshFullIndex()
}
