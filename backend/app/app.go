package app

import (
	"github.com/blue-monads/turnix/backend/app/server"
)

type App struct {
	happ   *HeadLess
	server *server.Server
}
