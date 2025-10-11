package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/engine"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
)

type Server struct {
	ctrl   *actions.Controller
	router *gin.Engine
	signer *signer.Signer

	engine *engine.Engine

	opt Option
}

type Option struct {
	Port        int
	Ctrl        *actions.Controller
	Signer      *signer.Signer
	Engine      *engine.Engine
	Host        string
	GlobalJS    string
	SiteName    string
	LocalSocket string
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
	err = s.listenUnixSocket()
	if err != nil {
		return err
	}

	s.router.Run(fmt.Sprintf(":%d", s.opt.Port))

	return nil
}

func (s *Server) listenUnixSocket() error {

	pp.Println("@listen_unix_socket", s.opt.LocalSocket)

	if s.opt.LocalSocket != "" {
		return nil
	}

	// delete old socket

	os.Remove(s.opt.LocalSocket)

	l, err := net.Listen("unix", s.opt.LocalSocket)
	if err != nil {
		log.Println("listen_unix_socket error:", err.Error())
		return err
	}

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				log.Fatal("accept error:", err.Error())
				return
			}

			func(c net.Conn) {
				defer c.Close()

				out, err := json.Marshal(map[string]any{
					"port": s.opt.Port,
					"host": s.opt.Host,
				})

				if err != nil {
					log.Fatal("json marshal error:", err.Error())
					return
				}

				_, err = c.Write(out)
				if err != nil {
					log.Fatal("Write: ", err)
				}

			}(c)

		}

	}()

	return nil
}
