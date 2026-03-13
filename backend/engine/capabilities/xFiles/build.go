package cfiles

import (
	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

var (
	Name         = "xFiles"
	Icon         = `<i class="fa-solid fa-folder-open"></i>`
	OptionFields = []xcapability.CapabilityOptionField{}
)

func init() {
	b := xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			appTyped := app.(xtypes.App)
			return &FilesBuilder{
				app: appTyped,
			}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	}

	registry.RegisterCapability(Name, b)
	registry.RegisterCapability("xfiles", b)
}

type FilesBuilder struct {
	app xtypes.App
}

func (b *FilesBuilder) Build(handle xcapability.XCapabilityHandle) (xcapability.Capability, error) {

	model := handle.GetModel()
	fileOps := b.app.Database().GetFileOps()

	return &FilesCapability{
		fileOps:   fileOps,
		installId: model.InstallID,
		handle:    handle,
	}, nil
}

func (b *FilesBuilder) Serve(ctx *gin.Context) {}

func (b *FilesBuilder) Name() string {
	return Name
}

func (b *FilesBuilder) GetDebugData() map[string]any {
	return map[string]any{}
}
