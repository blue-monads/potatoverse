package backend

import (
	"log/slog"

	"github.com/blue-monads/turnix/backend/app"
	"github.com/blue-monads/turnix/backend/services/datahub/database"
	"github.com/blue-monads/turnix/backend/services/signer"
)

type Options struct {
	DBFile string
}

func NewNoHead(options Options) *app.HeadLess {

	logger := slog.Default()

	db, err := database.NewDB(options.DBFile, logger)
	if err != nil {
		logger.Error("Failed to initialize database", "err", err)
		return nil
	}

	app := app.NewHeadLess(app.Option{
		Database: db,
		Logger:   logger,
		Signer:   signer.New([]byte("default-signer-key")),
	})

	return app
}
