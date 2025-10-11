package backend

import (
	"log/slog"
	"path"

	"github.com/blue-monads/turnix/backend/app"
	"github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/services/datahub/database"
	"github.com/blue-monads/turnix/backend/services/mailer/stdio"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/xtypes"
)

type Options struct {
	DBFile string
	Port   int
	SeedDB bool
	Host   string
	Name   string
}

func NewNoHead(options Options) (*app.HeadLess, error) {

	logger := slog.Default()

	db, err := database.NewDB(options.DBFile, logger)
	if err != nil {
		logger.Error("Failed to initialize database", "err", err)
		return nil, err
	}

	masterSecret := "default-master-secret"

	m := stdio.NewMailer(logger.With("module", "mailer"))

	if options.Name == "" {
		options.Name = "PotatoVerse"
	}

	app := app.NewHeadLess(app.Option{
		Database: db,
		Logger:   logger,
		Signer:   signer.New([]byte(masterSecret)),
		AppOpts: &xtypes.AppOptions{
			Port:         options.Port,
			Host:         options.Host,
			MasterSecret: masterSecret,
			Debug:        true,
			WorkingDir:   "./tmp",
			Name:         options.Name,
		},
		Mailer: m,
	})

	return app, nil
}

func NewDevApp(options Options) (*app.App, error) {

	happ, err := NewNoHead(options)
	if err != nil {
		return nil, err
	}

	if options.SeedDB {
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

func NewProdApp(config *xtypes.AppOptions) (*app.App, error) {
	happ, err := NewNoHead(Options{
		DBFile: path.Join(config.WorkingDir, "data.db"),
		Port:   config.Port,
		SeedDB: false,
		Host:   config.Host,
		Name:   config.Name,
	})
	if err != nil {
		return nil, err
	}

	return app.NewApp(happ), nil
}
