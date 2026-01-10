package corehub

import (
	"github.com/blue-monads/potatoverse/backend/services/corehub/buddyhub"
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/services/sockd"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

type CoreHub struct {
	app   xtypes.App
	db    datahub.Database
	sockd *sockd.Sockd

	signer *signer.Signer

	buddyhub *buddyhub.BuddyHub
}

func NewCoreHub(app xtypes.App) *CoreHub {
	db := app.Database()

	sockd := app.Sockd().(*sockd.Sockd)

	return &CoreHub{
		app:    app,
		db:     db,
		sockd:  sockd,
		signer: app.Signer(),
	}
}

func (c *CoreHub) Run() error {

	logger := c.app.Logger().With("module", "corehub")

	c.buddyhub = buddyhub.NewBuddyHub(buddyhub.Options{
		Logger: logger,
		App:    c.app,
	})

	return nil
}

func (c *CoreHub) GetBuddyHub() *buddyhub.BuddyHub {
	return c.buddyhub
}
