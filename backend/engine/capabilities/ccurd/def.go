package ccurd

import (
	"fmt"
	"regexp"

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

	methods := map[string]*Methods{}

	if err := opts.AsJson(&methods); err != nil {
		return nil, err
	}

	for _, method := range methods {
		for _, validator := range method.Validators {
			if validator.RegexPattern != "" {
				cr, err := regexp.Compile(validator.RegexPattern)
				if err != nil {
					return nil, err
				}

				validator.compiledRegex = cr
			}
		}
	}

	return &PingCapability{
		db:      b.app.Database().GetLowCapabilityDBOps(fmt.Sprint(spaceId)),
		signer:  b.app.Signer(),
		methods: methods,
		spaceId: spaceId,
	}, nil
}

func (b *CcurdBuilder) Serve(ctx *gin.Context) {}

func (p *CcurdBuilder) Name() string {
	return Name
}
