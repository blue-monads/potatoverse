package registry

import (
	"maps"
	"sync"

	"github.com/blue-monads/turnix/backend/engine/addons"
)

var (
	AddOnBuilderFactories = make(map[string]addons.BuilderFactory)
	mLock                 = sync.RWMutex{}
)

func RegisterAddOn(name string, factory addons.BuilderFactory) {
	AddOnBuilderFactories[name] = factory
}

func GetAddOnBuilderFactories() (map[string]addons.BuilderFactory, error) {
	mLock.RLock()
	defer mLock.RUnlock()

	copy := make(map[string]addons.BuilderFactory)
	maps.Copy(copy, AddOnBuilderFactories)

	return copy, nil
}
