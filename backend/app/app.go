package app

import (
	"log/slog"

	controller "github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/app/server"
	"github.com/blue-monads/turnix/backend/engine"
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/signer"
)

type App struct {
	happ   *HeadLess
	server *server.Server
}

func NewApp(happ *HeadLess) *App {
	return &App{
		happ: happ,
		server: server.NewServer(server.Option{
			Port:   happ.AppOpts.Port,
			Ctrl:   happ.Controller().(*controller.Controller),
			Signer: happ.Signer(),
			Engine: happ.Engine(),
		}),
	}
}

func (a *App) Init() error {
	return a.happ.Init()
}

func (a *App) Start() error {

	err := a.happ.Start()
	if err != nil {
		return err
	}

	return a.server.Start()
}

func (a *App) Database() datahub.Database {
	return a.happ.db
}

func (a *App) Signer() *signer.Signer {
	return a.happ.signer
}

func (a *App) Logger() *slog.Logger {
	return a.happ.logger
}

func (a *App) Controller() any {
	return a.happ.ctrl
}

func (a *App) Engine() *engine.Engine {
	return a.happ.engine
}
