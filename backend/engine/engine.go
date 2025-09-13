package engine

import (
	"sync"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/gin-gonic/gin"
)

type indexItem struct {
	packageId int64
	spaceId   int64
}

type Engine struct {
	db           datahub.Database
	RoutingIndex map[string]indexItem
	riLock       sync.RWMutex
}

func NewEngine(db datahub.Database) *Engine {
	return &Engine{
		db:           db,
		RoutingIndex: make(map[string]indexItem),
	}
}

func (e *Engine) LoadRoutingIndex() error {

	nextRoutingIndex := make(map[string]indexItem)

	spaces, err := e.db.ListSpaces()
	if err != nil {
		return err
	}

	for _, space := range spaces {
		if space.OwnsNamespace {
			nextRoutingIndex[space.NamespaceKey] = indexItem{
				packageId: space.PackageID,
				spaceId:   space.ID,
			}
		}
	}

	e.riLock.Lock()
	e.RoutingIndex = nextRoutingIndex
	e.riLock.Unlock()

	return nil
}

func (e *Engine) ServeSpaceFile(ctx *gin.Context) {

	spaceKey := ctx.Param("space_key")

	e.riLock.RLock()
	ri, ok := e.RoutingIndex[spaceKey]
	e.riLock.RUnlock()
	if !ok {
		ctx.JSON(404, gin.H{"error": "space not found"})
		return
	}

	e.db.GetPackageFileStreaming(ri.packageId, ri.spaceId, ctx.Writer)

}

func (e *Engine) ServePluginFile(ctx *gin.Context) {

}

func (e *Engine) SpaceApi(ctx *gin.Context) {

}

func (e *Engine) PluginApi(ctx *gin.Context) {

}
