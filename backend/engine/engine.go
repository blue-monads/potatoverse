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
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/models"
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
)

type Engine struct {
	db            datahub.Database
	RoutingIndex  map[string]*SpaceRouteIndexItem
	riLock        sync.RWMutex
	workingFolder string

	addons AddOnHub

	runtime Runtime

	logger *slog.Logger

	app xtypes.App
}

func NewEngine(db datahub.Database, workingFolder string) *Engine {
	e := &Engine{
		db:            db,
		workingFolder: workingFolder,
		RoutingIndex:  make(map[string]*SpaceRouteIndexItem),
		runtime: Runtime{
			execs:     make(map[int64]*luaz.Luaz),
			execsLock: sync.RWMutex{},
		},
		logger: slog.Default().With("module", "engine"),
		addons: AddOnHub{
			goodies:  make(map[string]xtypes.AddOn),
			glock:    sync.RWMutex{},
			builders: make(map[string]xtypes.AddOnBuilder),
		},
	}

	e.addons.parent = e

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

	go e.runtime.cleanupExecs()

	return e.LoadRoutingIndex()
}

func (e *Engine) ServeSpaceFile(ctx *gin.Context) {

	pp.Println("@ServeSpaceFile/1")

	spaceKey := ctx.Param("space_key")
	spaceId := extractDomainSpaceId(ctx.Request.URL.Host)

	pp.Println("@ServeSpaceFile/3")

	sIndex := e.getIndex(spaceKey, spaceId)

	if sIndex == nil {
		keys := make([]string, 0)
		for key := range e.RoutingIndex {
			keys = append(keys, key)
		}

		pp.Println("@ServeSpaceFile/4", keys)
		pp.Println("@ServeSpaceFile/4")
		httpx.WriteErrString(ctx, "space not found")
		return
	}

	pp.Println("@ServeSpaceFile/5")

	switch sIndex.routeOption.RouterType {
	case "simple", "":
		e.serveSimpleRoute(ctx, sIndex)
	case "dynamic":
		pp.Println("@ServeSpaceFile/6")
		e.serveDynamicRoute(ctx, sIndex)
	default:
		httpx.WriteErrString(ctx, "router type not supported")
		return
	}

}

func (e *Engine) ServePluginFile(ctx *gin.Context) {

}

func (e *Engine) ServeAddon(ctx *gin.Context) {
	spaceKey := ctx.Param("space_key")
	addonName := ctx.Param("addon_name")

	spaceId := extractDomainSpaceId(ctx.Request.URL.Host)

	index := e.getIndex(spaceKey, spaceId)

	if index == nil {
		httpx.WriteErr(ctx, errors.New("space not found"))
		return
	}

	e.addons.Handle(spaceId, addonName, ctx)

}

func (e *Engine) ServeAddonRoot(ctx *gin.Context) {
	addonName := ctx.Param("addon_name")
	e.addons.HandleRoot(addonName, ctx)
}

func (e *Engine) SpaceApi(ctx *gin.Context) {

	spaceKey := ctx.Param("space_key")
	spaceId := int64(0)

	if matches := spaceIdPattern.FindStringSubmatch(ctx.Request.URL.Host); matches != nil {
		sid, _ := strconv.ParseInt(matches[1], 10, 64)
		spaceId = sid
	}

	sIndex := e.getIndex(spaceKey, spaceId)

	if sIndex == nil {
		ctx.JSON(404, gin.H{"error": "space not found"})
		return
	}

	e.runtime.ExecHttp(spaceKey, sIndex.installedId, sIndex.spaceId, ctx)

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

func (e *Engine) SpaceInfo(nsKey string) (*SpaceInfo, error) {

	ri := e.getIndex(nsKey, 0)

	if ri == nil {
		return nil, errors.New("space not found")
	}

	space, err := e.db.GetSpaceOps().GetSpace(ri.spaceId)
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

// private

// s-12.example.com
var spaceIdPattern = regexp.MustCompile(`^s-(\d+)\.`)

func extractDomainSpaceId(domain string) int64 {
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

	pp.Println("@ropt", ropt)
	pp.Println("@name", name)
	pp.Println("@path", path)

	return name, path
}

func (e *Engine) GetAddonHub() *AddOnHub {
	return &e.addons
}
