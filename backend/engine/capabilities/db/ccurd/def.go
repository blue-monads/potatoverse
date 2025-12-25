package ccurd

import (
	"fmt"
	"regexp"

	"github.com/blue-monads/turnix/backend/engine/registry"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/lazydata"
	"github.com/blue-monads/turnix/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

var (
	Name         = "ccurd"
	Icon         = ""
	OptionFields = []xcapability.CapabilityOptionField{
		{
			Name:        "Methods",
			Key:         "methods",
			Description: "Define the methods to use",
			Type:        "object",
			Default:     "{}",
		},
	}
)

func init() {

	registry.RegisterCapability(Name, xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			appTyped := app.(xtypes.App)
			return &CcurdBuilder{app: appTyped}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	})
}

type CcurdBuilder struct {
	app xtypes.App
}

func (b *CcurdBuilder) Build(handle xcapability.XCapabilityHandle) (xcapability.Capability, error) {
	model := handle.GetModel()

	methods, err := LoadMethods(handle.GetOptionsAsLazyData())
	if err != nil {
		return nil, err
	}

	return &CcurdCapability{
		db:           b.app.Database().GetLowCapabilityDBOps(fmt.Sprint(model.SpaceID)),
		signer:       b.app.Signer(),
		methods:      methods,
		spaceId:      model.SpaceID,
		installId:    model.InstallID,
		capabilityId: model.ID,
		engine:       b.app.Engine().(xtypes.Engine),
	}, nil
}

func (b *CcurdBuilder) GetDebugData() map[string]any {
	return map[string]any{}
}

type CcurdOptions struct {
	Methods map[string]*Methods `json:"methods"`
}

func LoadMethods(opts lazydata.LazyData) (map[string]*Methods, error) {

	optsData := CcurdOptions{}
	if err := opts.AsJson(&optsData); err != nil {
		return nil, err
	}

	for _, method := range optsData.Methods {
		for _, validator := range method.Validators {
			if validator.Regex != "" {
				cr, err := regexp.Compile(validator.Regex)
				if err != nil {
					return nil, err
				}

				validator.compiledRegex = cr
			}
		}
	}

	return optsData.Methods, nil
}

func (b *CcurdBuilder) Serve(ctx *gin.Context) {}

func (p *CcurdBuilder) Name() string {
	return Name
}
