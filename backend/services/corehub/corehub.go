package corehub

import (
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/sockd"
	"github.com/blue-monads/turnix/backend/xtypes"
)

type CoreHub struct {
	db    datahub.Database
	sockd *sockd.Sockd
}

func NewCoreHub(app xtypes.App) *CoreHub {
	db := app.Database()

	sockd := app.Sockd().(*sockd.Sockd)

	return &CoreHub{
		db:    db,
		sockd: sockd,
	}
}

func (c *CoreHub) Run() error {

	return nil
}
