package registry

import (
	"maps"
	"sync"

	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/xcapability"
)

var (
	registryStore = &RegistryStore{
		capabilityBuilderFactories: make(map[string]xcapability.CapabilityBuilderFactory),
		executorBuilderFactories:   make(map[string]xtypes.ExecutorBuilderFactory),
		rootExecutorFactories:      make(map[string]xtypes.RootExecutorFactory),
		rsLock:                     sync.Mutex{},
	}
)

type RegistryStore struct {
	capabilityBuilderFactories map[string]xcapability.CapabilityBuilderFactory
	executorBuilderFactories   map[string]xtypes.ExecutorBuilderFactory
	rootExecutorFactories      map[string]xtypes.RootExecutorFactory
	rsLock                     sync.Mutex
}

func (rs *RegistryStore) RegisterCapabilityBuilderFactory(name string, factory xcapability.CapabilityBuilderFactory) {
	rs.rsLock.Lock()
	defer rs.rsLock.Unlock()

	rs.capabilityBuilderFactories[name] = factory
}

func (rs *RegistryStore) GetCapabilityBuilderFactories() map[string]xcapability.CapabilityBuilderFactory {
	rs.rsLock.Lock()
	defer rs.rsLock.Unlock()

	copy := make(map[string]xcapability.CapabilityBuilderFactory)
	maps.Copy(copy, rs.capabilityBuilderFactories)
	return copy
}

func (rs *RegistryStore) RegisterExecutorBuilderFactory(name string, factory xtypes.ExecutorBuilderFactory) {
	rs.rsLock.Lock()
	defer rs.rsLock.Unlock()

	rs.executorBuilderFactories[name] = factory
}

func (rs *RegistryStore) GetExecutorBuilderFactories() map[string]xtypes.ExecutorBuilderFactory {
	rs.rsLock.Lock()
	defer rs.rsLock.Unlock()

	copy := make(map[string]xtypes.ExecutorBuilderFactory)
	maps.Copy(copy, rs.executorBuilderFactories)
	return copy
}

func (rs *RegistryStore) RegisterRootExecutorFactory(name string, factory xtypes.RootExecutorFactory) {
	rs.rsLock.Lock()
	defer rs.rsLock.Unlock()

	rs.rootExecutorFactories[name] = factory
}

func (rs *RegistryStore) GetRootExecutorFactories() map[string]xtypes.RootExecutorFactory {
	rs.rsLock.Lock()
	defer rs.rsLock.Unlock()

	copy := make(map[string]xtypes.RootExecutorFactory)
	maps.Copy(copy, rs.rootExecutorFactories)
	return copy
}
