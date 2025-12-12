package corehub

import (
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/sockd"
)

type CoreHub struct {
	db    datahub.Database
	sockd *sockd.Sockd
}

func NewCoreHub(db datahub.Database, sockd *sockd.Sockd) *CoreHub {
	return &CoreHub{
		db:    db,
		sockd: sockd,
	}
}

func (c *CoreHub) Run() error {

	return nil
}
