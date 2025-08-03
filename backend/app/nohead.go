package app

import (
	"log/slog"

	"github.com/blue-monads/turnix/backend/services/datahub"
)

type HeadLessOptions struct {
	Database datahub.Database
	Logger   *slog.Logger
}

// headless means it has no http server attached to it
type HeadLess struct {
	db     datahub.Database
	logger *slog.Logger
}

func NewHeadLess(options HeadLessOptions) *HeadLess {
	return &HeadLess{
		db:     options.Database,
		logger: options.Logger,
	}
}
