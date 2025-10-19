package registry

import (
	"maps"
	"sync"

	"github.com/blue-monads/turnix/backend/engine/addons"
)

var (
	AddOnBuilders = make(map[string]addons.BuilderFactory)
	mLock         = sync.RWMutex{}
)

func RegisterBuilder(name string, factory addons.BuilderFactory) {
	AddOnBuilders[name] = factory
}

func GetAddOnBuilder() (map[string]addons.BuilderFactory, error) {
	mLock.RLock()
	defer mLock.RUnlock()

	copy := make(map[string]addons.BuilderFactory)
	maps.Copy(copy, AddOnBuilders)

	return copy, nil
}
