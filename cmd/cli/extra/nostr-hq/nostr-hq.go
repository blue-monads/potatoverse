package nostrhq

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/blue-monads/potatoverse/cmd/cli"
	"github.com/fiatjaf/relayer/v2"
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

	r := Relay{}

	server, err := relayer.NewServer(&r)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}
	if err := server.Start("0.0.0.0", port); err != nil {
		log.Fatalf("server terminated: %v", err)
	}

	return nil

}
