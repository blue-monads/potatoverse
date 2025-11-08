package app

import (
	"log/slog"

	"github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/app/server"
	"github.com/blue-monads/turnix/backend/engine"
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/xtypes"
)

var _ xtypes.App = (*App)(nil)

type App struct {
	happ   *HeadLess
	server *server.Server
}

func NewApp(happ *HeadLess) *App {

	hosts := make([]string, len(happ.AppOpts.Hosts))
	for i, host := range happ.AppOpts.Hosts {
		hosts[i] = host.Name
	}

	return &App{
		happ: happ,
		server: server.NewServer(server.Option{
			Port:        happ.AppOpts.Port,
			Ctrl:        happ.Controller().(*actions.Controller),
			Signer:      happ.Signer(),
			Engine:      happ.Engine().(*engine.Engine),
			Hosts:       hosts,
			LocalSocket: happ.AppOpts.SocketFile,
			SiteName:    "Demo",
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

func (a *App) Engine() any {

	return a.happ.engine
}

func (a *App) Config() any {
	return a.happ.AppOpts
}
