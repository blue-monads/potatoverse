package xugroup

import (
	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

var (
	Name         = "xUgroup"
	Icon         = ""
	OptionFields = []xcapability.CapabilityOptionField{
		{
			Name: "All User Groups",
			Key:  "all_user_group",
			Type: "boolean",
		},
		{
			Name: "User Groups",
			Key:  "ugroups",
			Type: "multi_select",
		},
	}
)

func init() {
	registry.RegisterCapability(Name, xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			appTyped := app.(xtypes.App)
			return &UgroupBuilder{app: appTyped}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	})
}

type UgroupBuilder struct {
	app xtypes.App
}

func (b *UgroupBuilder) Build(handle xcapability.XCapabilityHandle) (xcapability.Capability, error) {
	model := handle.GetModel()

	opts := &UgroupOptions{}
	handle.GetOptions(opts)

	return &UgroupCapability{
		app:          b.app,
		handle:       handle,
		spaceId:      model.SpaceID,
		installId:    model.InstallID,
		allUserGroup: opts.AllUserGroup,
		ugroups:      opts.Ugroups,
	}, nil
}

func (b *UgroupBuilder) Serve(ctx *gin.Context) {}

func (b *UgroupBuilder) Name() string {
	return Name
}

func (b *UgroupBuilder) GetDebugData() map[string]any {
	return map[string]any{}
}
