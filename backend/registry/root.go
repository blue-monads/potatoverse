package registry

import (
	"github.com/blue-monads/turnix/backend/xtypes"
)

func RegisterRootExecutorFactory(name string, factory xtypes.RootExecutorFactory) {
	registryStore.RegisterRootExecutorFactory(name, factory)
}

func GetRootExecutorFactories() map[string]xtypes.RootExecutorFactory {
	return registryStore.GetRootExecutorFactories()
}
