package server

import (
	"fmt"

	controller "github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/gin-gonic/gin"
)

type Server struct {
	ctrl   *controller.Controller
	router *gin.Engine
	signer *signer.Signer
	port   int
}

type Option struct {
	Port   int
	Ctrl   *controller.Controller
	Signer *signer.Signer
}

func NewServer(opt Option) *Server {
	return &Server{
		ctrl:   opt.Ctrl,
		signer: opt.Signer,
		port:   opt.Port,
	}
}

func (s *Server) Start() error {

	s.router = gin.Default()

	s.bindRoutes()

	s.router.Run(fmt.Sprintf(":%d", s.port))

	return nil
}
