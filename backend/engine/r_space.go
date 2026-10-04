package engine

import (
	"maps"
	"path"
	"strings"
	"sync"
	"time"

	xutils "github.com/blue-monads/potatoverse/backend/utils"
	"github.com/blue-monads/potatoverse/backend/utils/libx/httpx"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes/models"
	"github.com/gin-gonic/gin"
)

type SpaceRouter struct {
	engine *Engine

	RoutingIndex map[string]*SpaceRouteIndexItem
	riLock       sync.RWMutex

	reloadPackageIds chan int64
	fullReload       chan struct{}
	stopEloop        chan struct{}
	stopOnce         sync.Once
}

func NewSpaceRouter(engine *Engine) *SpaceRouter {
	return &SpaceRouter{
		engine:           engine,
		RoutingIndex:     make(map[string]*SpaceRouteIndexItem),
		riLock:           sync.RWMutex{},
		reloadPackageIds: make(chan int64, 20),
		fullReload:       make(chan struct{}, 1),
		stopEloop:        make(chan struct{}),
	}
}

func (r *SpaceRouter) Start() {
	go r.startEloop()
	r.LoadRoutingIndex()
}

func (r *SpaceRouter) Close() {
	r.stopOnce.Do(func() {
		close(r.stopEloop)
	})
}

func (r *SpaceRouter) GetRoutingIndexCopy() map[string]*SpaceRouteIndexItem {
	indexCopy := make(map[string]*SpaceRouteIndexItem)
	r.riLock.RLock()
	maps.Copy(indexCopy, r.RoutingIndex)
	r.riLock.RUnlock()
	return indexCopy
}

func (r *SpaceRouter) ServeSpaceFile(ctx *gin.Context, spaceId int64) {

	qq.Println("@ServeSpaceFile/1")

	var spaceKey string

	if spaceId == 0 {
		spaceKey = ctx.Param("space_key")
		spaceId = xutils.ExtractSpaceId(ctx.Request.Host)
	}

	qq.Println("@ServeSpaceFile/3")

	sIndex := r.getIndexRetry(spaceKey, spaceId)

	if sIndex == nil {

		keys := make([]string, 0)
		r.riLock.RLock()
		for key := range r.RoutingIndex {
			keys = append(keys, key)
		}
		r.riLock.RUnlock()

		qq.Println("@ServeSpaceFile/4", keys)
		qq.Println("@ServeSpaceFile/4")
		httpx.WriteErrString(ctx, "space not found")
		return
	}

	qq.Println("@ServeSpaceFile/5")

	switch sIndex.routeOption.RouterType {
	case "simple", "":
		r.serveSimpleRoute(ctx, sIndex)
	case "dynamic":
		qq.Println("@ServeSpaceFile/6")
		r.serveDynamicRoute(ctx, sIndex)
	default:
		httpx.WriteErrString(ctx, "router type not supported")
		return
	}

}

func (r *SpaceRouter) Serve(ctx *gin.Context) {
	r.ServeSpaceFile(ctx, 0)
}

func (r *SpaceRouter) GetIndex(spaceKey string, spaceId int64) *SpaceRouteIndexItem {
	return r.getIndex(spaceKey, spaceId)
}

func (r *SpaceRouter) GetIndexRetry(spaceKey string, spaceId int64) *SpaceRouteIndexItem {
	return r.getIndexRetry(spaceKey, spaceId)
}

func buildPackageFilePath(filePath string, ropt *models.PotatoRouteOptions) (string, string) {
	nameParts := strings.Split(filePath, "/")
	name := nameParts[len(nameParts)-1]
	pathParts := nameParts[:len(nameParts)-1]

	ppath := strings.Join(pathParts, "/")
	ppath = path.Join(ropt.ServeFolder, ppath)

	ppath = strings.TrimLeft(ppath, "/")

	if ropt.TrimPathPrefix != "" {
		ppath = strings.TrimPrefix(ppath, ropt.TrimPathPrefix)
	}

	if ropt.ForceIndexHtmlFile && name == "" {
		name = "index.html"
	}

	if ropt.ForceHtmlExtension && !strings.Contains(name, ".") {
		name = name + ".html"
	}

	qq.Println("@ropt", ropt)
	qq.Println("@name", name)
	qq.Println("@path", ppath)

	return name, ppath
}

func (r *SpaceRouter) startEloop() {

	readAllPendingPackageIds := func() []int64 {
		spaceIds := make([]int64, 0)

	main:
		for {

			select {
			case sid := <-r.reloadPackageIds:
				spaceIds = append(spaceIds, sid)
			default:
				break main
			}
		}
		return spaceIds
	}

	pendingFullReload := func() bool {
		for {
			select {
			case <-r.fullReload:
				return true
			default:
				return false
			}
		}
	}

	sTimer := time.NewTicker(time.Second * 2)
	defer sTimer.Stop()

	for {
		select {
		case <-r.stopEloop:
			return
		case <-sTimer.C:
			if pendingFullReload() {
				r.loadRoutingIndex()
				readAllPendingPackageIds()
				continue
			}

			packageIds := readAllPendingPackageIds()
			if len(packageIds) > 0 {
				r.loadRoutingIndexForPackages(packageIds...)
			}
		}
	}

}
