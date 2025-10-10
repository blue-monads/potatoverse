package server

import (
	"fmt"

	"github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/engine"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/gin-gonic/gin"
)

type Server struct {
	ctrl   *actions.Controller
	router *gin.Engine
	signer *signer.Signer

	engine *engine.Engine

	opt Option
}

type Option struct {
	Port     int
	Ctrl     *actions.Controller
	Signer   *signer.Signer
	Engine   *engine.Engine
	Host     string
	GlobalJS string
	SiteName string
}

func NewServer(opt Option) *Server {
	return &Server{
		ctrl:   opt.Ctrl,
		signer: opt.Signer,
		engine: opt.Engine,
		opt:    opt,
	}
}

func (s *Server) Start() error {
	err := s.buildGlobalJS()
	if err != nil {
		return err
	}

	s.router = gin.Default()

	s.bindRoutes()

	s.router.Run(fmt.Sprintf(":%d", s.opt.Port))

	return nil
}
