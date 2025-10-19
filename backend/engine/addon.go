package engine

import (
	"errors"
	"fmt"
	"sync"

	"github.com/blue-monads/turnix/backend/engine/addons"
	"github.com/blue-monads/turnix/backend/engine/registry"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/gin-gonic/gin"
)

type AddOnHub struct {
	parent  *Engine
	goodies map[string]addons.AddOn
	glock   sync.RWMutex

	builders map[string]addons.Builder
}

func (gh *AddOnHub) Init() error {
	app := gh.parent.app

	app.Logger().Info("Initializing AddOnHub")

	builderFactories, err := registry.GetAddOnBuilderFactories()
	if err != nil {
		return err
	}

	gh.builders = make(map[string]addons.Builder)
	for name, factory := range builderFactories {
		builder, err := factory(app)
		if err != nil {
			return err
		}

		gh.builders[name] = builder

	}

	app.Logger().Info("AddOnHub initialized")

	return nil
}

func (gh *AddOnHub) Handle(spaceId int64, name string, ctx *gin.Context) {
	gs, err := gh.get(name, spaceId)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	gs.Handle(ctx)
}

func (gh *AddOnHub) HandleRoot(name string, ctx *gin.Context) {
	builder, ok := gh.builders[name]
	if !ok {
		httpx.WriteErr(ctx, errors.New("addon builder not found"))
		return
	}

	builder.Serve(ctx)
}

func (gh *AddOnHub) List(spaceId int64) ([]string, error) {
	keys := make([]string, 0)

	for key := range gh.builders {
		keys = append(keys, key)
	}

	return keys, nil
}

func (gh *AddOnHub) GetMeta(spaceId int64, gname, method string) (map[string]any, error) {
	gs, err := gh.get(gname, spaceId)
	if err != nil {
		return nil, err
	}

	return gs.GetMeta(method)
}

func (gh *AddOnHub) Execute(spaceId int64, gname, method string, params addons.LazyData) (map[string]any, error) {
	gs, err := gh.get(gname, spaceId)
	if err != nil {
		return nil, err
	}

	return gs.Execute(method, params)
}

// private

func (gh *AddOnHub) get(name string, spaceId int64) (addons.AddOn, error) {
	key := fmt.Sprintf("%s:%d", name, spaceId)

	gh.glock.RLock()
	gs, ok := gh.goodies[key]
	gh.glock.RUnlock()

	if !ok {
		gbFactory, ok := gh.builders[name]
		if !ok {
			return nil, errors.New("goodies builder not found")
		}

		instance, err := gbFactory.Build(spaceId)
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
