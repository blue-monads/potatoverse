package engine

import (
	"sync"

	"github.com/blue-monads/turnix/backend/services/datahub"
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
