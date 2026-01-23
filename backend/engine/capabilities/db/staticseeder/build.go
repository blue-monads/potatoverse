package staticseeder

import (
	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

var (
	Name         = "staticseeder"
	Icon         = ""
	OptionFields = []xcapability.CapabilityOptionField{
		{
			Name:        "Seed Folder",
			Key:         "staticseeder_folder",
			Description: "Path to the seed folder containing JSON files (e.g., 'seed' or 'db/seed')",
			Type:        "text",
			Default:     "seed",
			Required:    true,
		},
	}
)

func init() {
	registry.RegisterCapability(Name, xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			appTyped := app.(xtypes.App)
			return &StaticSeederBuilder{app: appTyped}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	})
}

type StaticSeederBuilder struct {
	app xtypes.App
}

func (b *StaticSeederBuilder) Name() string {
	return Name
}

func (b *StaticSeederBuilder) Build(handle xcapability.XCapabilityHandle) (xcapability.Capability, error) {
	model := handle.GetModel()

	opts := handle.GetOptionsAsLazyData()

	seedFolder := opts.GetFieldAsString("staticseeder_folder")
	if seedFolder == "" {
		seedFolder = "seed"
	}

	db := b.app.Database().GetLowPackageDBOps(model.InstallID)

	capability := &StaticSeederCapability{
		seedFolder:   seedFolder,
		builder:      b,
		installId:    model.InstallID,
		spaceId:      model.SpaceID,
		capabilityId: model.ID,
		db:           db,
	}

	return capability, nil
}

func (b *StaticSeederBuilder) Serve(ctx *gin.Context) {}

func (b *StaticSeederBuilder) GetDebugData() map[string]any {
	return map[string]any{
		"name": Name,
	}
}
