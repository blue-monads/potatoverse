package engine

import (
	"errors"
	"maps"
	"strings"
	"sync"

	"github.com/blue-monads/turnix/backend/engine/executors/luaz"
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
)

type Engine struct {
	db            datahub.Database
	RoutingIndex  map[string]SpaceRouteIndexItem
	riLock        sync.RWMutex
	workingFolder string

	runtime Runtime

	app xtypes.App
}

func NewEngine(db datahub.Database, workingFolder string) *Engine {
	return &Engine{
		db:            db,
		workingFolder: workingFolder,
		RoutingIndex:  make(map[string]SpaceRouteIndexItem),
		runtime: Runtime{
			execs:     make(map[int64]*luaz.Luaz),
			execsLock: sync.RWMutex{},
		},
	}
}

func (e *Engine) GetDebugData() map[string]any {
	indexCopy := make(map[string]SpaceRouteIndexItem)
	e.riLock.RLock()
	maps.Copy(indexCopy, e.RoutingIndex)
	e.riLock.RUnlock()

	return map[string]any{
		"runtime_data":  e.runtime.GetDebugData(),
		"routing_index": indexCopy,
	}

}

func (e *Engine) LoadRoutingIndex() error {

	nextRoutingIndex := make(map[string]SpaceRouteIndexItem)

	spaces, err := e.db.ListSpaces()
	if err != nil {
		return err
	}

	for _, space := range spaces {
		if space.OwnsNamespace {
			nextRoutingIndex[space.NamespaceKey] = SpaceRouteIndexItem{
				packageId: space.PackageID,
				spaceId:   space.ID,
				routeOption: RouteOption{
					ServeFolder:        "public",
					TrimPathPrefix:     "",
					ForceHtmlExtension: false,
					ForceIndexHtmlFile: false,
				},
			}
		}
	}

	e.riLock.Lock()
	e.RoutingIndex = nextRoutingIndex
	e.riLock.Unlock()

	return nil
}

func (e *Engine) Start(app xtypes.App) error {
	e.app = app
	e.runtime.parent = e

	go e.runtime.cleanupExecs()

	return e.LoadRoutingIndex()
}

func (e *Engine) ServeSpaceFile(ctx *gin.Context) {

	spaceKey := ctx.Param("space_key")

	e.riLock.RLock()
	sIndex, ok := e.RoutingIndex[spaceKey]
	e.riLock.RUnlock()
	if !ok {
		ctx.JSON(404, gin.H{"error": "space not found"})
		return
	}

	filePath := ctx.Param("files")

	name, path := buildPackageFilePath(filePath, &sIndex.routeOption)

	pp.Println("@name", name)
	pp.Println("@path", path)

	if name == "" {
		name = "index.html"
	}

	fmeta, err := e.db.GetPackageFileMetaByPath(sIndex.packageId, path, name)
	if err != nil {
		ctx.JSON(404, gin.H{"error": "file not found"})
		return
	}

	e.db.GetPackageFileStreaming(sIndex.packageId, fmeta.ID, ctx.Writer)

}

func (e *Engine) ServePluginFile(ctx *gin.Context) {

}

func (e *Engine) SpaceApi(ctx *gin.Context) {

	spaceKey := ctx.Param("space_key")

	e.riLock.RLock()
	ri, ok := e.RoutingIndex[spaceKey]
	e.riLock.RUnlock()
	if !ok {
		ctx.JSON(404, gin.H{"error": "space not found"})
		return
	}

	e.runtime.ExecHttp(spaceKey, ri.packageId, ri.spaceId, ctx)

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

	e.riLock.RLock()
	ri, ok := e.RoutingIndex[nsKey]
	e.riLock.RUnlock()
	if !ok {
		return nil, errors.New("space not found")
	}

	space, err := e.db.GetSpace(ri.spaceId)
	if err != nil {
		return nil, err
	}

	pkg, err := e.db.GetPackage(space.PackageID)
	if err != nil {
		return nil, err
	}

	return &SpaceInfo{
		ID:            space.ID,
		NamespaceKey:  space.NamespaceKey,
		OwnsNamespace: space.OwnsNamespace,
		PackageName:   pkg.Name,
		PackageInfo:   pkg.Info,
	}, nil

}

func buildPackageFilePath(filePath string, ropt *RouteOption) (string, string) {
	nameParts := strings.Split(filePath, "/")
	name := nameParts[len(nameParts)-1]
	pathParts := nameParts[:len(nameParts)-1]
	pathParts = append(pathParts, ropt.ServeFolder)

	path := strings.Join(pathParts, "/")
	path = strings.TrimLeft(path, "/")

	if ropt.TrimPathPrefix != "" {
		path = strings.TrimPrefix(path, ropt.TrimPathPrefix)
	}

	if ropt.ForceHtmlExtension && !strings.Contains(name, ".") {
		name = name + ".html"
	}

	if ropt.ForceIndexHtmlFile && name == "" {
		name = "index.html"
	}

	return name, path
}
