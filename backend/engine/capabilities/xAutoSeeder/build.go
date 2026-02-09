package autoseeder

import (
	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

var (
	Name         = "autoseeder"
	Icon         = `<i class="fa-solid fa-fill"></i>`
	OptionFields = []xcapability.CapabilityOptionField{
		{
			Name:        "Seed Struct File",
			Key:         "autoseeder_struct_file",
			Description: "Path to the autoseeder.struct.json file (e.g., 'autoseeder.struct.json' or 'db/autoseeder.struct.json')",
			Type:        "text",
			Default:     "autoseeder.struct.json",
			Required:    true,
		},
	}
)

func init() {
	registry.RegisterCapability(Name, xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			appTyped := app.(xtypes.App)
			return &AutoseederBuilder{app: appTyped}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	})
}

type AutoseederBuilder struct {
	app xtypes.App
}

func (b *AutoseederBuilder) Name() string {
	return Name
}

func (b *AutoseederBuilder) Build(handle xcapability.XCapabilityHandle) (xcapability.Capability, error) {
	model := handle.GetModel()

	opts := handle.GetOptionsAsLazyData()

	autoseederStructFile := opts.GetFieldAsString("autoseeder_struct_file")
	if autoseederStructFile == "" {
		autoseederStructFile = "autoseeder.struct.json"
	}

	db := b.app.Database().GetLowPackageDBOps(model.InstallID)

	capability := &AutoseederCapability{
		autoseederStructFile: autoseederStructFile,
		builder:              b,
		installId:            model.InstallID,
		spaceId:              model.SpaceID,
		capabilityId:         model.ID,
		db:                   db,
	}

	return capability, nil
}

func (b *AutoseederBuilder) Serve(ctx *gin.Context) {}

func (b *AutoseederBuilder) GetDebugData() map[string]any {
	return map[string]any{
		"name": Name,
	}
}
