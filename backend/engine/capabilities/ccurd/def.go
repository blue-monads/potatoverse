package ccurd

import (
	"github.com/blue-monads/turnix/backend/engine/registry"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
)

var (
	Name         = "ccurd"
	Icon         = ""
	OptionFields = []xtypes.CapabilityOptionField{
		{
			Name:        "Add Radom number to the result",
			Key:         "add_random_number",
			Description: "Add a random number to the result",
			Type:        "boolean",
			Default:     "false",
		},
	}
)

func init() {

	registry.RegisterCapability(Name, xtypes.CapabilityBuilderFactory{
		Builder: func(app xtypes.App) (xtypes.CapabilityBuilder, error) {
			return &CcurdBuilder{app: app}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	})
}

type CcurdBuilder struct {
	app xtypes.App
}

func (b *CcurdBuilder) Build(spaceId int64, opts xtypes.LazyData) (xtypes.Capability, error) {
	return &PingCapability{
		app:     b.app,
		spaceId: spaceId,
	}, nil
}

func (b *CcurdBuilder) Serve(ctx *gin.Context) {}

func (p *CcurdBuilder) Name() string {
	return "ping"
}
