package registry

import (
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
)

func RegisterCapability(name string, factory xcapability.CapabilityBuilderFactory) {
	registryStore.RegisterCapabilityBuilderFactory(name, factory)
}

func GetCapabilityBuilderFactories() (map[string]xcapability.CapabilityBuilderFactory, error) {
	return registryStore.GetCapabilityBuilderFactories(), nil
}
