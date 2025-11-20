package registry

import (
	"maps"
	"sync"

	"github.com/blue-monads/turnix/backend/xtypes"
)

var (
	ExecutorBuilderFactories = make(map[string]xtypes.ExecutorBuilderFactory)
	meLock                   = sync.RWMutex{}
)

func RegisterExecutor(name string, factory xtypes.ExecutorBuilderFactory) {
	ExecutorBuilderFactories[name] = factory
}

func GetExecutorBuilderFactories() (map[string]xtypes.ExecutorBuilderFactory, error) {
	meLock.RLock()
	defer meLock.RUnlock()

	copy := make(map[string]xtypes.ExecutorBuilderFactory)
	maps.Copy(copy, ExecutorBuilderFactories)

	return copy, nil
}
