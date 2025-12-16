package registry

import (
	"maps"
	"sync"

	"github.com/blue-monads/turnix/backend/xtypes/xcapability"
)

var (
	CapabilityBuilderFactories = make(map[string]xcapability.CapabilityBuilderFactory)
	mLock                      = sync.RWMutex{}
)

func RegisterCapability(name string, factory xcapability.CapabilityBuilderFactory) {
	CapabilityBuilderFactories[name] = factory
}

func GetCapabilityBuilderFactories() (map[string]xcapability.CapabilityBuilderFactory, error) {
	mLock.RLock()
	defer mLock.RUnlock()

	copy := make(map[string]xcapability.CapabilityBuilderFactory)
	maps.Copy(copy, CapabilityBuilderFactories)

	return copy, nil
}
