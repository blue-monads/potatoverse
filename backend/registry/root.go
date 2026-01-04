package registry

import (
	"maps"
	"sync"

	"github.com/blue-monads/turnix/backend/xtypes"
)

var (
	RootExecutorFactories = make(map[string]xtypes.RootExecutorFactory)
	reLock                = sync.RWMutex{}
)

func RegisterRootExecutorFactory(name string, factory xtypes.RootExecutorFactory) {
	RootExecutorFactories[name] = factory
}

func GetRootExecutorFactories() map[string]xtypes.RootExecutorFactory {
	reLock.RLock()
	defer reLock.RUnlock()

	copy := make(map[string]xtypes.RootExecutorFactory)
	maps.Copy(copy, RootExecutorFactories)

	return copy
}
