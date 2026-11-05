package server

import (
	"log/slog"
	"os"
	"testing"

	"github.com/blue-monads/potatoverse/backend/app/actions"
	rtbuddy "github.com/blue-monads/potatoverse/backend/app/server/rt_buddy"
	"github.com/blue-monads/potatoverse/backend/engine"
	"github.com/blue-monads/potatoverse/backend/services/buddyhub"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
)

func TestBindRoutes_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	tmpDir, err := os.MkdirTemp("", "potatoverse_route_test_*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	opts := &xtypes.AppOptions{
		Name:         "PotatoTest",
		Port:         8080,
		WorkingDir:   tmpDir,
		MasterSecret: "super_secret_master_key_123456",
	}

	logger := slog.Default()
	bhub := buddyhub.NewBuddyHub(opts, logger)
	sig := signer.New([]byte(opts.MasterSecret))

	eng := engine.NewEngine(engine.EngineOption{
		Logger:        logger,
		WorkingFolder: tmpDir,
	})

	s := &Server{
		router:      router,
		ctrl:        actions.New(actions.Option{Logger: logger}),
		signer:      sig,
		engine:      eng,
		buddyRoutes: rtbuddy.New(bhub, 8080),
		opt: Option{
			Logger:   logger,
			BuddyHub: bhub,
			Signer:   sig,
			Engine:   eng,
		},
	}

	// This must not panic (e.g. wildcard route conflict)
	s.bindRoutes()
}
