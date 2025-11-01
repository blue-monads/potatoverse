package engine

import (
	"errors"
	"fmt"
	"sync"

	"github.com/blue-monads/turnix/backend/engine/registry"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type CapabilityHub struct {
	parent  *Engine
	goodies map[string]xtypes.Capability
	glock   sync.RWMutex

	builders map[string]xtypes.CapabilityBuilder
}

func (gh *CapabilityHub) Init() error {
	app := gh.parent.app

	app.Logger().Info("Initializing CapabilityHub")

	builderFactories, err := registry.GetCapabilityBuilderFactories()
	if err != nil {
		return err
	}

	gh.builders = make(map[string]xtypes.CapabilityBuilder)
	for name, factory := range builderFactories {
		builder, err := factory(app)
		if err != nil {
			return err
		}

		gh.builders[name] = builder

	}

	app.Logger().Info("CapabilityHub initialized")

	return nil
}

func (gh *CapabilityHub) Handle(spaceId int64, name string, ctx *gin.Context) {
	gs, err := gh.get(name, spaceId)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	gs.Handle(ctx)
}

func (gh *CapabilityHub) HandleRoot(name string, ctx *gin.Context) {
	builder, ok := gh.builders[name]
	if !ok {
		httpx.WriteErr(ctx, errors.New("capability builder not found"))
		return
	}

	builder.Serve(ctx)
}

func (gh *CapabilityHub) List(spaceId int64) ([]string, error) {
	keys := make([]string, 0)

	for key := range gh.builders {
		keys = append(keys, key)
	}

	return keys, nil
}

func (gh *CapabilityHub) Methods(spaceId int64, gname string) ([]string, error) {
	gs, err := gh.get(gname, spaceId)
	if err != nil {
		return nil, err
	}

	return gs.ListActions()
}

func (gh *CapabilityHub) GetMeta(spaceId int64, gname, method string) (map[string]any, error) {
	gs, err := gh.get(gname, spaceId)
	if err != nil {
		return nil, err
	}

	return gs.GetActionMeta(method)
}

func (gh *CapabilityHub) Execute(spaceId int64, gname, method string, params xtypes.LazyData) (map[string]any, error) {
	gs, err := gh.get(gname, spaceId)
	if err != nil {
		return nil, err
	}

	return gs.ExecuteAction(method, params)
}

// private

func (gh *CapabilityHub) get(name string, spaceId int64) (xtypes.Capability, error) {
	key := fmt.Sprintf("%s:%d", name, spaceId)

	gh.glock.RLock()
	gs, ok := gh.goodies[key]
	gh.glock.RUnlock()

	if !ok {
		gbFactory, ok := gh.builders[name]
		if !ok {
			return nil, errors.New("capability builder not found")
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

/*

CapabilityResolver

-- user group
-- system
-- install




*/
