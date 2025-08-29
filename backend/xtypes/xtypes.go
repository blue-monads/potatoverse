package xtypes

import (
	"log/slog"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/signer"
)

type App interface {
	Init() error
	Start() error
	Database() datahub.Database
	Signer() *signer.Signer
	Logger() *slog.Logger
	Controller() any
}
