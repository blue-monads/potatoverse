package registry

import (
	"maps"
	"sync"

	"github.com/blue-monads/turnix/backend/xtypes"
)

var (
	CapabilityBuilderFactories = make(map[string]xtypes.CapabilityBuilderFactory)
	mLock                      = sync.RWMutex{}
)

func RegisterCapability(name string, factory xtypes.CapabilityBuilderFactory) {
	CapabilityBuilderFactories[name] = factory
}

func GetCapabilityBuilderFactories() (map[string]xtypes.CapabilityBuilderFactory, error) {
	mLock.RLock()
	defer mLock.RUnlock()

	copy := make(map[string]xtypes.CapabilityBuilderFactory)
	maps.Copy(copy, CapabilityBuilderFactories)

	return copy, nil
}
