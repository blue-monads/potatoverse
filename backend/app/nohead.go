package app

import (
	"log/slog"

	"github.com/blue-monads/turnix/backend/app/controller"
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/signer"
)

type Option struct {
	Database datahub.Database
	Logger   *slog.Logger
	Signer   *signer.Signer
}

// headless means it has no http server attached to it
type HeadLess struct {
	db     datahub.Database
	signer *signer.Signer
	logger *slog.Logger
	ctrl   *controller.Controller
}

func NewHeadLess(opt Option) *HeadLess {
	return &HeadLess{
		db:     opt.Database,
		signer: opt.Signer,
		logger: opt.Logger,
		ctrl: controller.New(controller.Option{
			Database: opt.Database,
			Logger:   opt.Logger,
			Signer:   opt.Signer,
		}),
	}
}
