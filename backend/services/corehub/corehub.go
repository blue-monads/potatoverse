package corehub

import (
	"path"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/sockd"
	"github.com/blue-monads/turnix/backend/xtypes"
)

type CoreHub struct {
	db    datahub.Database
	sockd *sockd.Sockd

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

	return nil
}
