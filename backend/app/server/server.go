package server

import (
	"github.com/blue-monads/turnix/backend/app/controller"
	"github.com/gin-gonic/gin"
)

type Server struct {
	ctrl   *controller.Controller
	router *gin.Engine
}

func NewServer(ctrl *controller.Controller) *Server {
	return &Server{
		ctrl: ctrl,
	}
}

func (s *Server) Start() error {

	s.router = gin.Default()

	s.router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	s.router.Run(":8080")

	return nil
}
