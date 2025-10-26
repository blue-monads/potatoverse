package backend

import (
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/blue-monads/turnix/backend/app"
	"github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/services/datahub/database"
	"github.com/blue-monads/turnix/backend/services/mailer/stdio"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/xtypes"
)

func BuildApp(options *xtypes.AppOptions, seedDB bool) (*app.App, error) {

	logger := slog.Default()

	db, err := database.NewDB(fmt.Sprintf("%s/data.db", options.WorkingDir), logger)
	if err != nil {
		logger.Error("Failed to initialize database", "err", err)
		return nil, err
	}

	m := stdio.NewMailer(logger.With("module", "mailer"))

	if options.Name == "" {
		options.Name = "PotatoVerse"
	}

	happ := app.NewHeadLess(app.Option{
		Database: db,
		Logger:   logger,
		Signer:   signer.New([]byte(options.MasterSecret)),
		AppOpts: &xtypes.AppOptions{
			Port:         options.Port,
			Hosts:        options.Hosts,
			MasterSecret: options.MasterSecret,
			Debug:        options.Debug,
			WorkingDir:   options.WorkingDir,
			Name:         options.Name,
			SocketFile:   options.SocketFile,
			Mailer:       options.Mailer,
		},
		Mailer: m,
	})

	if seedDB {
		ctrl := happ.Controller().(*actions.Controller)

		ugroups, err := ctrl.ListUserGroups()
		if err != nil {
			return nil, err
		}

		if len(ugroups) == 0 {
			err = ctrl.AddUserGroup("admin", "Admin group")
			if err != nil {
				return nil, err
			}

			err = ctrl.AddUserGroup("normal", "Normal group")
			if err != nil {
				return nil, err
			}

			_, err = ctrl.AddAdminUserDirect("demo", "demogodTheGreat_123", "demo@example.com")
			if err != nil {
				return nil, err
			}
		}

	}

	return app.NewApp(happ), nil
}

func NewDevApp(config *xtypes.AppOptions, seedDB bool) (*app.App, error) {
	if config.WorkingDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		config.WorkingDir = path.Join(cwd, ".pdata")
	}

	if config.MasterSecret == "" {
		config.MasterSecret = "default-master-secret"
	}

	if config.SocketFile == "" {
		config.SocketFile = path.Join(config.WorkingDir, "./potatoverse.sock")
	}

	app, err := BuildApp(config, seedDB)
	if err != nil {
		return nil, err
	}

	return app, nil
}

func NewProdApp(config *xtypes.AppOptions, seedDB bool) (*app.App, error) {
	app, err := BuildApp(config, seedDB)
	if err != nil {
		return nil, err
	}

	return app, nil

}
