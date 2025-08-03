package controller

import (
	"log/slog"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/xtypes"
)

type Option struct {
	Database datahub.Database
	Logger   *slog.Logger
	Signer   *signer.Signer
	AppOpts  *xtypes.AppOptions
}

type Controller struct {
	database datahub.Database
	logger   *slog.Logger
	signer   *signer.Signer
	AppOpts  *xtypes.AppOptions
}

func New(opt Option) *Controller {
	return &Controller{
		database: opt.Database,
		logger:   opt.Logger,
		signer:   opt.Signer,
	}
}
