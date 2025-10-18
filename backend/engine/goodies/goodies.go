package goodies

import (
	"errors"
	"fmt"
	"sync"

	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type LazyData interface {
	AsMap() (map[string]any, error)
	AsJson(target any) error
}

type Goodies interface {
	Name() string
	Handle(ctx *gin.Context) error
	List() ([]string, error)
	GetMeta(name string) (map[string]any, error)
	Execute(method string, params LazyData) (map[string]any, error)
}

type GoodiesBuilderFactory func(app xtypes.App) (GoodiesBuilder, error)

type GoodiesBuilder func(spaceId int64) (Goodies, error)

type GoodiesHub struct {
	goodies map[string]Goodies
	glock   sync.RWMutex
	app     xtypes.App

	builders map[string]GoodiesBuilder
}

func (gh *GoodiesHub) Handle(spaceId int64, name string, ctx *gin.Context) error {
	gs, err := gh.get(name, spaceId)
	if err != nil {
		return err
	}

	return gs.Handle(ctx)
}

func (gh *GoodiesHub) List(spaceId int64) ([]string, error) {
	keys := make([]string, 0)

	for key := range gh.builders {
		keys = append(keys, key)
	}

	return keys, nil
}

func (gh *GoodiesHub) GetMeta(spaceId int64, gname, method string) (map[string]any, error) {
	gs, err := gh.get(gname, spaceId)
	if err != nil {
		return nil, err
	}

	return gs.GetMeta(method)
}

func (gh *GoodiesHub) Execute(spaceId int64, gname, method string, params LazyData) (map[string]any, error) {
	gs, err := gh.get(gname, spaceId)
	if err != nil {
		return nil, err
	}

	return gs.Execute(method, params)
}

// private

func (gh *GoodiesHub) get(name string, spaceId int64) (Goodies, error) {
	key := fmt.Sprintf("%s:%d", name, spaceId)

	gh.glock.RLock()
	gs, ok := gh.goodies[key]
	gh.glock.RUnlock()

	if !ok {
		gbFactory, ok := gh.builders[name]
		if !ok {
			return nil, errors.New("goodies builder not found")
		}

		instance, err := gbFactory(spaceId)
		if err != nil {
			return nil, err
		}

		gh.glock.Lock()
		defer gh.glock.Unlock()
		oldInstance, ok := gh.goodies[key]
		if ok && oldInstance != nil {
			return oldInstance, nil
		}

		gh.goodies[key] = instance

		return instance, nil
	}

	return gs, nil
}
