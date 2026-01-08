package corehub

import (
	"path"

	"github.com/blue-monads/turnix/backend/services/corehub/buddyhub"
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/sockd"
	"github.com/blue-monads/turnix/backend/xtypes"
)

type CoreHub struct {
	app   xtypes.App
	db    datahub.Database
	sockd *sockd.Sockd

	buddyhub *buddyhub.BuddyHub

	buddyDir string
}

func NewCoreHub(app xtypes.App) *CoreHub {
	db := app.Database()

	sockd := app.Sockd().(*sockd.Sockd)
	appConfig := app.Config().(*xtypes.AppOptions)

	buddyDir := path.Join(appConfig.WorkingDir, "buddy")

	return &CoreHub{
		db:       db,
		sockd:    sockd,
		buddyDir: buddyDir,
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
