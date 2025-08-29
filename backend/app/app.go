package app

import (
	"github.com/blue-monads/turnix/backend/app/server"
)

type App struct {
	happ   *HeadLess
	server *server.Server
}

func NewApp(happ *HeadLess) *App {
	return &App{
		happ:   happ,
		server: server.NewServer(happ.Controller()),
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
