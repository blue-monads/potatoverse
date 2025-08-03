package controller

import (
	"log/slog"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/signer"
)

type Option struct {
	Database datahub.Database
	Logger   *slog.Logger
	Signer   *signer.Signer
}

type Controller struct {
	database datahub.Database
	logger   *slog.Logger
	signer   *signer.Signer
}

func New(opt Option) *Controller {
	return &Controller{
		database: opt.Database,
		logger:   opt.Logger,
		signer:   opt.Signer,
	}
}
