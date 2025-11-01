package registry

import (
	"maps"
	"sync"

	"github.com/blue-monads/turnix/backend/xtypes"
)

var (
	AddOnBuilderFactories = make(map[string]xtypes.AddOnBuilderFactory)
	mLock                 = sync.RWMutex{}
)

func RegisterAddOn(name string, factory xtypes.AddOnBuilderFactory) {
	AddOnBuilderFactories[name] = factory
}

func GetAddOnBuilderFactories() (map[string]xtypes.AddOnBuilderFactory, error) {
	mLock.RLock()
	defer mLock.RUnlock()

	copy := make(map[string]xtypes.AddOnBuilderFactory)
	maps.Copy(copy, AddOnBuilderFactories)

	return copy, nil
}
