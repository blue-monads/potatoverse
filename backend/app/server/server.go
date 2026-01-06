package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/engine"
	"github.com/blue-monads/turnix/backend/engine/hubs/buddyhub"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/gin-gonic/gin"
)

type Server struct {
	ctrl     *actions.Controller
	router   *gin.Engine
	signer   *signer.Signer
	buddyhub *buddyhub.BuddyHub

	engine *engine.Engine

	opt Option
}

type Option struct {
	Port        int
	Ctrl        *actions.Controller
	Signer      *signer.Signer
	Engine      *engine.Engine
	Hosts       []string
	GlobalJS    string
	SiteName    string
	LocalSocket string

	// ServerKey just some identifier for the server, (lowercase a-z and numbers)
	// it could be hash for public key if node is tunneling traffic for other nodes

	ServerKey string
}

func NewServer(opt Option) *Server {
	if opt.ServerKey == "" {
		opt.ServerKey = "main"
	}

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

	s.buddyhub = s.engine.GetBuddyHub().(*buddyhub.BuddyHub)

	go func() {

		time.Sleep(2 * time.Second)

		fmt.Println("Server started:")
		fmt.Println("Listening on:\t\t", fmt.Sprintf("http://localhost:%d/zz/pages", s.opt.Port))
		fmt.Println("Node Pubkey:\t\t", s.buddyhub.GetPubkey())

	}()

	return s.router.Run(fmt.Sprintf(":%d", s.opt.Port))

}

func (s *Server) listenUnixSocket() error {

	qq.Println("@listen_unix_socket", s.opt.LocalSocket)

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
					"host": s.opt.Hosts,
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
