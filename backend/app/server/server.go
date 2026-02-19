package server

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/aurowora/compress"
	"github.com/blue-monads/potatoverse/backend/app/actions"
	rtbuddy "github.com/blue-monads/potatoverse/backend/app/server/rt_buddy"
	"github.com/blue-monads/potatoverse/backend/engine"
	"github.com/blue-monads/potatoverse/backend/services/buddyhub"
	"github.com/blue-monads/potatoverse/backend/services/corehub"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	limits "github.com/gin-contrib/size"
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

	BuddyHub *buddyhub.BuddyHub
	CoreHub  *corehub.CoreHub
	Logger   *slog.Logger
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

	s.buddyRoutes = rtbuddy.New(s.opt.BuddyHub, s.opt.Port)

	err := s.buildGlobalJS()
	if err != nil {
		return err
	}

	s.router = gin.Default()

	enableCOmpression :=
		os.Getenv("FRONTEND_DEV_SERVER") == "" &&
			os.Getenv("POTATO_DEV_SPACES") == ""

	if enableCOmpression {
		s.router.Use(compress.Compress(
			compress.WithAlgo(compress.DEFLATE, true),
			compress.WithAlgo(compress.GZIP, true),
			compress.WithAlgo(compress.ZSTD, true),
			compress.WithAlgo(compress.BROTLI, true),
		))
	}

	s.router.Use(limits.RequestSizeLimiter(100 * 1024 * 1024)) // 100 mb

	s.router.Use(s.buddyRoutes.BuddyAutoRouteMW())

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
			pubkey := s.opt.BuddyHub.GetPubkey()
			fmt.Println("Server started:")
			fmt.Println("Listening on:\t\t", fmt.Sprintf("http://localhost:%d/zz/pages", s.opt.Port))
			fmt.Println("Node Pubkey:\t\t", pubkey)

			tdomain := s.opt.BuddyHub.GetHQTunnelDomain()

			if tdomain != "" {
				fmt.Println("HQ Tunnel :\t\t", fmt.Sprintf("http://%s/zz/pages", tdomain))
			}

		}

		// GetHQTunnelDomain

		err = s.opt.BuddyHub.Start()
		if err != nil {
			panic(err)
		}

	}()

	return s.router.Run(fmt.Sprintf(":%d", s.opt.Port))

}

func (s *Server) listenUnixSocket() error {

	qq.Println("@listen_unix_socket", s.opt.LocalSocket)

	if s.opt.LocalSocket == "" {
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
