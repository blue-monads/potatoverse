package server

import (
	controller "github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/gin-gonic/gin"
)

type Server struct {
	ctrl   *controller.Controller
	router *gin.Engine
	signer *signer.Signer
}

func NewServer(ctrl *controller.Controller, signer *signer.Signer) *Server {
	return &Server{
		ctrl:   ctrl,
		signer: signer,
	}
}

func (s *Server) Start() error {

	s.router = gin.Default()

	s.bindRoutes()

	s.router.Run(":8080")

	return nil
}
