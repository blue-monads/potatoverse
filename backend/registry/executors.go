package registry

import (
	"github.com/blue-monads/turnix/backend/xtypes"
)

func RegisterExecutorBuilderFactory(name string, factory xtypes.ExecutorBuilderFactory) {
	registryStore.RegisterExecutorBuilderFactory(name, factory)
}

func GetExecutorBuilderFactories() map[string]xtypes.ExecutorBuilderFactory {
	return registryStore.GetExecutorBuilderFactories()
}
