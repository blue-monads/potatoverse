package xremote

import (
	"github.com/blue-monads/potatoverse/backend/engine/hubs/remotehub"
	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

var (
	Name         = "xRemote"
	Icon         = `<i class="fa-solid fa-satellite-dish"></i>`
	OptionFields = []xcapability.CapabilityOptionField{
		{
			Name:        "Target Space ID",
			Key:         "target_space_id",
			Description: "ID of the target space to bind to (optional if using remote_ctx_token)",
			Type:        "number",
		},
		{
			Name:        "Target Install ID",
			Key:         "target_install_id",
			Description: "ID of the target install/package to bind to (optional)",
			Type:        "number",
		},
	}
)

func init() {
	registry.RegisterCapability(xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			appTyped := app.(xtypes.App)
			return &RemoteBuilder{app: appTyped}, nil
		},
		Name:             Name,
		Icon:             Icon,
		FreeFieldOptions: true,
		OptionFields:     OptionFields,
	})
}

type RemoteBuilder struct {
	app xtypes.App
}

func (b *RemoteBuilder) Name() string {
	return Name
}

func (b *RemoteBuilder) Serve(ctx *gin.Context) {}

func (b *RemoteBuilder) GetDebugData() map[string]any {
	return map[string]any{"name": Name}
}

func (b *RemoteBuilder) Build(handle xcapability.XCapabilityHandle) (xcapability.Capability, error) {
	opts := handle.GetOptionsAsLazyData()
	targetSpaceId := int64(opts.GetFieldAsInt("target_space_id"))
	if targetSpaceId == 0 {
		targetSpaceId = int64(opts.GetFieldAsInt("space_id"))
	}

	targetInstallId := int64(opts.GetFieldAsInt("target_install_id"))
	if targetInstallId == 0 {
		targetInstallId = int64(opts.GetFieldAsInt("remote_install_id"))
	}

	var targetVerId int64
	if targetInstallId != 0 {
		pkg, err := b.app.Database().GetPackageInstallOps().GetPackage(targetInstallId)
		if err == nil && pkg != nil {
			targetVerId = pkg.ActiveInstallID
		}
	} else if targetSpaceId != 0 {
		space, err := b.app.Database().GetSpaceOps().GetSpace(targetSpaceId)
		if err == nil && space != nil {
			targetInstallId = space.InstalledId
			pkg, err := b.app.Database().GetPackageInstallOps().GetPackage(space.InstalledId)
			if err == nil && pkg != nil {
				targetVerId = pkg.ActiveInstallID
			}
		}
	}

	engine := b.app.Engine().(xtypes.Engine)
	var rhub *remotehub.RemoteHub
	if getter, ok := engine.(interface{ GetRemoteHub() *remotehub.RemoteHub }); ok {
		rhub = getter.GetRemoteHub()
	}

	return &RemoteCapability{
		app:                     b.app,
		remoteHub:               rhub,
		capHandle:               handle,
		defaultSpaceId:          targetSpaceId,
		defaultPackageId:        targetInstallId,
		defaultPackageVersionId: targetVerId,
	}, nil
}
