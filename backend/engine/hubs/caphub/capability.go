package caphub

import (
	"errors"
	"fmt"
	"sync"

	"github.com/blue-monads/turnix/backend/engine/registry"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/lazydata"
	"github.com/blue-monads/turnix/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

var _ xcapability.CapabilityHub = (*CapabilityHub)(nil)

type CapabilityHub struct {
	app xtypes.App

	goodies map[string]xcapability.Capability
	glock   sync.RWMutex

	builders         map[string]xcapability.CapabilityBuilder
	builderFactories map[string]xcapability.CapabilityBuilderFactory
}

func NewCapabilityHub() *CapabilityHub {
	return &CapabilityHub{

		goodies:          make(map[string]xcapability.Capability),
		glock:            sync.RWMutex{},
		builders:         make(map[string]xcapability.CapabilityBuilder),
		builderFactories: make(map[string]xcapability.CapabilityBuilderFactory),
	}
}

func (gh *CapabilityHub) Init(app xtypes.App) error {
	gh.app = app

	builderFactories, err := registry.GetCapabilityBuilderFactories()
	if err != nil {
		return err
	}

	gh.builderFactories = builderFactories

	gh.builders = make(map[string]xcapability.CapabilityBuilder)
	for name, factory := range builderFactories {
		builder, err := factory.Builder(app)
		if err != nil {
			return err
		}

		gh.builders[name] = builder

	}

	app.Logger().Info("CapabilityHub initialized")

	return nil
}

func (gh *CapabilityHub) GetDebugData(name string) map[string]any {
	builder, ok := gh.builders[name]
	if !ok {
		return nil
	}

	return builder.GetDebugData()
}

func (gh *CapabilityHub) Reload(installId int64, spaceId int64, name string) error {
	db := gh.app.Database().GetSpaceOps()

	cap, err := db.GetSpaceCapability(installId, name)
	if err != nil {
		return err
	}

	gg, err := gh.get(name, installId, spaceId)
	if err != nil {
		return err
	}

	next, err := gg.Reload(cap)
	if err != nil {
		return err
	}

	gh.set(name, installId, spaceId, next)

	return nil
}

func (gh *CapabilityHub) Handle(installId, spaceId int64, name string, ctx *gin.Context) {
	gs, err := gh.get(name, installId, spaceId)
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

func (gh *CapabilityHub) Methods(installId, spaceId int64, gname string) ([]string, error) {
	gs, err := gh.get(gname, installId, spaceId)
	if err != nil {
		return nil, err
	}

	return gs.ListActions()
}

func (gh *CapabilityHub) Execute(installId, spaceId int64, gname, method string, params lazydata.LazyData) (any, error) {
	gs, err := gh.get(gname, installId, spaceId)
	if err != nil {
		return nil, err
	}

	return gs.Execute(method, params)
}

func (gh *CapabilityHub) Definations() []CapabilityDefination {
	definations := make([]CapabilityDefination, 0)
	for _, factory := range gh.builderFactories {
		definations = append(definations, CapabilityDefination{
			Name:         factory.Name,
			Icon:         factory.Icon,
			OptionFields: factory.OptionFields,
		})
	}
	return definations
}

type CapabilityDefination struct {
	Name         string                              `json:"name"`
	Icon         string                              `json:"icon"`
	OptionFields []xcapability.CapabilityOptionField `json:"option_fields"`
}

// private

func (gh *CapabilityHub) set(name string, installId, spaceId int64, i xcapability.Capability) {
	key := fmt.Sprintf("%s:%d", name, spaceId)

	gh.glock.Lock()
	gh.goodies[key] = i
	gh.glock.Unlock()

}

func (gh *CapabilityHub) get(name string, installId, spaceId int64) (xcapability.Capability, error) {
	key := fmt.Sprintf("%s:%d", name, spaceId)

	gh.glock.RLock()
	gs, ok := gh.goodies[key]
	gh.glock.RUnlock()

	if !ok {
		gbFactory, ok := gh.builders[name]
		if !ok {
			return nil, errors.New("capability builder not found")
		}

		db := gh.app.Database().GetSpaceOps()

		cap, err := db.GetSpaceCapability(installId, name)
		if err != nil {
			return nil, err
		}

		handle := NewCapabilityHandle(gh.app, cap)
		instance, err := gbFactory.Build(handle)
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
