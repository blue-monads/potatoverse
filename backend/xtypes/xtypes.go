package xtypes

import (
	"log/slog"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/signer"
)

type App interface {
	ExecId() string
	Init() error
	Start() error
	Database() datahub.Database
	Signer() *signer.Signer
	Logger() *slog.Logger
	Controller() any
	Engine() any
	Config() any
	Sockd() any
	CoreHub() any
}
