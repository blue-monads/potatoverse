package funnelhq

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/buddyhub/funnel"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/cmd/cli"
	"github.com/gin-gonic/gin"
)

func init() {
	cli.RegisterExtraCommand("nostr-hq", func(args []string) error {
		return Run(context.Background())
	})
}

func Run(ctx context.Context) error {

	port := 7447

	portStr := os.Getenv("POTATO_HQ_PORT")
	if portStr != "" {
		porti, err := strconv.Atoi(portStr)
		if err != nil {
			log.Fatalf("failed to parse port: %v", err)
		}

		port = porti
	}

	// start funnel server

	funnel := funnel.New()
	gin.SetMode(gin.TestMode)
	router := gin.New()

	qq.Println("@main/2")

	router.GET("/funnel/register/:serverId", func(c *gin.Context) {

		qq.Println("@main/2.1{SERVER_ID}", c.Param("serverId"))

		serverId := c.Param("serverId")

		qq.Println("@main/2.2{SERVER_ID}", serverId)

		funnel.HandleServerWebSocket(serverId, c)
	})

	qq.Println("@main/3")

	// Route endpoint
	router.NoRoute(func(c *gin.Context) {
		// http://serverid.localhost/path

		url := c.Request.Host

		qq.Println("@main/3.1{URL}", url)

		serverId := strings.Split(url, ".")[0]

		qq.Println("@main/3.2{SERVER_ID}", serverId)

		funnel.HandleRoute(serverId, c)
	})

	qq.Println("@main/4")

	router.Run(fmt.Sprintf(":%d", port))

	return nil
}
