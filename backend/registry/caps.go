package registry

import (
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
)

func RegisterCapability(factory xcapability.CapabilityBuilderFactory) {
	registryStore.RegisterCapabilityBuilderFactory(factory.Name, factory)
}

func GetCapabilityBuilderFactories() (map[string]xcapability.CapabilityBuilderFactory, error) {
	return registryStore.GetCapabilityBuilderFactories(), nil
}
