package backend

import (
	"log/slog"

	"github.com/blue-monads/turnix/backend/app"
	"github.com/blue-monads/turnix/backend/services/datahub/database"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/xtypes"
)

type Options struct {
	DBFile string
	PORT   int
	SeedDB bool
}

func NewNoHead(options Options) (*app.HeadLess, error) {

	logger := slog.Default()

	db, err := database.NewDB(options.DBFile, logger)
	if err != nil {
		logger.Error("Failed to initialize database", "err", err)
		return nil, err
	}

	masterSecret := "default-master-secret"

	app := app.NewHeadLess(app.Option{
		Database: db,
		Logger:   logger,
		Signer:   signer.New([]byte(masterSecret)),
		AppOpts: &xtypes.AppOptions{
			Name:         "Turnix",
			Port:         options.PORT,
			Host:         "localhost",
			MasterSecret: masterSecret,
			Debug:        true,
			WorkingDir:   "./tmp",
		},
	})

	return app, nil
}

func NewApp(options Options) (*app.App, error) {

	happ, err := NewNoHead(options)
	if err != nil {
		return nil, err
	}

	if options.SeedDB {

		happ.Controller().AddAdminUserDirect("demo", "demogodTheGreat_123", "demo@example.com")

	}

	return app.NewApp(happ), nil
}
