package startup

import (
	"github.com/blue-monads/potatoverse/backend/app"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

func NewProdApp(config *xtypes.AppOptions, seedDB bool) (*app.App, error) {
	app, err := BuildApp(config, seedDB)
	if err != nil {
		return nil, err
	}

	return app, nil

}
