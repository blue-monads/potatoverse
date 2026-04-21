package startup

import (
	"os"
	"path"

	"github.com/blue-monads/potatoverse/backend/app"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

func NewDevApp(config *xtypes.AppOptions, seedDB bool) (*app.App, error) {
	if config.WorkingDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		config.WorkingDir = path.Join(cwd, "pdata")
	}

	if config.MasterSecret == "" {
		config.MasterSecret = "default-master-secret"
	}

	if config.SocketFile == "" {
		config.SocketFile = path.Join(config.WorkingDir, "./potatoverse.sock")
	}

	if len(config.Repos) == 0 {
		config.Repos = repohub.Default

	}

	app, err := BuildApp(config, seedDB)
	if err != nil {
		return nil, err
	}

	return app, nil
}
