package xremotedb

import (
	"errors"

	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

var (
	Name         = "xRemoteDB"
	Icon         = `<i class="fa-solid fa-database"></i>`
	OptionFields = []xcapability.CapabilityOptionField{}
)

func init() {
	registry.RegisterCapability(Name, xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			appTyped := app.(xtypes.App)
			return &RemoteDbBuilder{app: appTyped}, nil
		},
		Name:             Name,
		Icon:             Icon,
		FreeFieldOptions: true,
		OptionFields:     OptionFields,
	})
}

type RemoteDbBuilder struct {
	app xtypes.App
}

func (b *RemoteDbBuilder) Name() string {
	return Name
}

func (b *RemoteDbBuilder) Serve(ctx *gin.Context) {}

func (b *RemoteDbBuilder) GetDebugData() map[string]any {
	return map[string]any{"name": Name}
}

func (b *RemoteDbBuilder) Build(handle xcapability.XCapabilityHandle) (xcapability.Capability, error) {
	opts := handle.GetOptionsAsLazyData()
	installedId := opts.GetFieldAsInt("remote_install_id")
	if installedId == 0 {
		return nil, errors.New("remote_install_id is required")
	}

	db := b.app.Database().GetLowPackageDBOps(int64(installedId))

	return &RemoteDbCapability{
		installId: int64(installedId),
		db:        db,
		capHandle: handle,
	}, nil
}
