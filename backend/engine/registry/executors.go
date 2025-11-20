package registry

import (
	"maps"
	"sync"

	"github.com/blue-monads/turnix/backend/xtypes"
)

var (
	executorBuilderFactories = make(map[string]xtypes.ExecutorBuilderFactory)
	meLock                   = sync.Mutex{}
)

func RegisterExecutorBuilderFactory(name string, factory xtypes.ExecutorBuilderFactory) {
	meLock.Lock()
	defer meLock.Unlock()

	executorBuilderFactories[name] = factory
}

func GetExecutorBuilderFactories() map[string]xtypes.ExecutorBuilderFactory {
	meLock.Lock()
	defer meLock.Unlock()

	copy := make(map[string]xtypes.ExecutorBuilderFactory)
	maps.Copy(copy, executorBuilderFactories)

	return copy
}
