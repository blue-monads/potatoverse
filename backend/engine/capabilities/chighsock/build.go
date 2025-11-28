package chighsock

import (
	"github.com/blue-monads/turnix/backend/engine/registry"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/services/sockd"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
)

var (
	Name         = "chighsock"
	Icon         = "socket"
	OptionFields = []xtypes.CapabilityOptionField{}
)

var (
	OKResponse = map[string]any{"success": true}
)

func init() {
	registry.RegisterCapability(Name, xtypes.CapabilityBuilderFactory{
		Builder: func(app xtypes.App) (xtypes.CapabilityBuilder, error) {
			return &ChighsockBuilder{app: app}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	})
}

type ChighsockBuilder struct {
	app xtypes.App
}

func (b *ChighsockBuilder) Build(model *dbmodels.SpaceCapability) (xtypes.Capability, error) {

	hs := b.app.Sockd().(*sockd.Sockd).GetHigher()

	return &ChighsockCapability{
		app:          b.app,
		spaceId:      model.SpaceID,
		installId:    model.InstallID,
		capabilityId: model.ID,
		signer:       b.app.Signer(),
		higher:       hs,
	}, nil
}

func (b *ChighsockBuilder) Serve(ctx *gin.Context) {}

func (b *ChighsockBuilder) Name() string {
	return Name
}
