package relayhttp

import (
	"sync"

	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

func init() {
	registry.RegisterCapability(Name, xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			return &RelayHttpBuilder{}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	})
}

type RelayHttpBuilder struct {
	httpRelays map[string]*RelayHttp
	rLock      sync.RWMutex
}

func (b *RelayHttpBuilder) Build(handle xcapability.XCapabilityHandle) (xcapability.Capability, error) {
	opts := handle.GetOptionsAsLazyData()
	allowSubKey := opts.GetFieldAsBool("allow_sub_key")

	return &RelayHttpCapability{
		parent:      b,
		handle:      handle,
		allowSubKey: allowSubKey,
	}, nil
}

func (b *RelayHttpBuilder) Serve(ctx *gin.Context) {}

func (b *RelayHttpBuilder) Name() string {
	return Name
}

func (b *RelayHttpBuilder) GetDebugData() map[string]any {
	return map[string]any{}
}
