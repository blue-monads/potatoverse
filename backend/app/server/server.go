package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/blue-monads/potatoverse/backend/app/actions"
	rtbuddy "github.com/blue-monads/potatoverse/backend/app/server/rt_buddy"
	"github.com/blue-monads/potatoverse/backend/engine"
	"github.com/blue-monads/potatoverse/backend/services/corehub"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/gin-gonic/gin"
)

type Server struct {
	ctrl   *actions.Controller
	router *gin.Engine
	signer *signer.Signer
	engine *engine.Engine
	opt    Option

	buddyRoutes *rtbuddy.BuddyRouteServer
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

	CoreHub *corehub.CoreHub

	ServerPubKey string
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
	pubkey := s.opt.CoreHub.GetBuddyHub().GetPubkey()
	s.opt.ServerPubKey = pubkey

	buddyhub := s.opt.CoreHub.GetBuddyHub()

	s.buddyRoutes = rtbuddy.New(buddyhub, s.opt.Port, s.opt.ServerPubKey)

	err := s.buildGlobalJS()
	if err != nil {
		return err
	}

	s.router = gin.Default()
	s.router.Use(s.buddyRoutes.BuddyAutoRouteMW)

	s.bindRoutes()
	err = s.listenUnixSocket()
	if err != nil {
		return err
	}

	existed := false

	defer func() {
		existed = true
	}()

	go func() {

		time.Sleep(2 * time.Second)

		if !existed {
			fmt.Println("Server started:")
			fmt.Println("Listening on:\t\t", fmt.Sprintf("http://localhost:%d/zz/pages", s.opt.Port))
			fmt.Println("Node Pubkey:\t\t", pubkey)
		}

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
