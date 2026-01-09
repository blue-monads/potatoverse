package actions

import (
	"log/slog"

	"github.com/blue-monads/potatoverse/backend/engine"
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/mailer"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

type Option struct {
	Database datahub.Database
	Logger   *slog.Logger
	Signer   *signer.Signer
	AppOpts  *xtypes.AppOptions
	Engine   *engine.Engine
	Mailer   mailer.Mailer
}

type Controller struct {
	database datahub.Database
	logger   *slog.Logger
	signer   *signer.Signer
	AppOpts  *xtypes.AppOptions
	engine   *engine.Engine
	mailer   mailer.Mailer
}

func New(opt Option) *Controller {
	return &Controller{
		database: opt.Database,
		logger:   opt.Logger,
		signer:   opt.Signer,
		AppOpts:  opt.AppOpts,
		engine:   opt.Engine,
		mailer:   opt.Mailer,
	}
}
