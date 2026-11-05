package cloudy

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/blue-monads/potatoverse/backend/app"
	"github.com/blue-monads/potatoverse/backend/app/actions"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub"
	"github.com/blue-monads/potatoverse/backend/services/buddyhub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database"
	"github.com/blue-monads/potatoverse/backend/services/mailer"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
	turso "turso.tech/database/tursogo"
)

type SubApp struct {
	Name   string
	App    xtypes.App
	Engine *gin.Engine

	tursoDB *turso.TursoSyncDb
	config  *Config
	mailer  mailer.Mailer

	mu        sync.Mutex
	ready     bool
	readyCh   chan struct{}
	readyOnce sync.Once
	startErr  error
}

func NewSubApp(ctx context.Context, config *Config, name string, bootstrap bool, m mailer.Mailer, remote tenantRemote) (*SubApp, error) {
	tenantDir := filepath.Join(config.WorkingDir, "tenants", name)
	if err := os.MkdirAll(tenantDir, 0o755); err != nil {
		return nil, err
	}

	tursoDB, err := turso.NewTursoSyncDb(ctx, turso.TursoSyncDbConfig{
		Path:             filepath.Join(tenantDir, "app.db"),
		RemoteUrl:        remote.URL,
		AuthToken:        remote.AuthToken,
		BootstrapIfEmpty: &bootstrap,
		Namespace:        remote.Namespace,
	})
	if err != nil {
		return nil, err
	}

	return &SubApp{
		Name:    name,
		tursoDB: tursoDB,
		config:  config,
		mailer:  m,
		readyCh: make(chan struct{}),
	}, nil
}

func (s *SubApp) Push(ctx context.Context) error {
	return s.tursoDB.Push(ctx)
}

func (s *SubApp) Load(ctx context.Context, adminName, adminPassword, adminEmail string) error {
	s.mu.Lock()
	if s.ready {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	if _, err := s.tursoDB.Pull(ctx); err != nil {
		log.Printf("tenant %s: initial pull: %v", s.Name, err)
	}

	db, err := s.tursoDB.Connect(ctx)
	if err != nil {
		return fmt.Errorf("tenant %s: connect: %w", s.Name, err)
	}

	logger := slog.Default().With("tenant", s.Name)

	adb, err := database.FromSqlHandle(db, logger)
	if err != nil {
		return fmt.Errorf("tenant %s: open database: %w", s.Name, err)
	}

	baseDomain := normalizeDomain(s.config.Domain)
	tenantHost := fmt.Sprintf("%s.%s", s.Name, baseDomain)
	workDir := filepath.Join(s.config.WorkingDir, "tenants", s.Name)

	appOpts := &xtypes.AppOptions{
		Port:         0,
		WorkingDir:   workDir,
		MasterSecret: s.config.MasterSecret,
		Name:         fmt.Sprintf("Cloudy/%s", s.Name),
		Repos:        repohub.Default,
		Hosts: []xtypes.Host{
			{Name: tenantHost},
			{Name: "*." + tenantHost},
		},
	}

	bhub := buddyhub.NewDummyBuddyHub()
	tenantEngine := gin.New()
	tenantEngine.Use(gin.Logger(), gin.Recovery())

	happ := app.New(app.Option{
		Database:          adb,
		Logger:            logger,
		Signer:            signer.New([]byte(s.config.MasterSecret)),
		AppOpts:           appOpts,
		Mailer:            s.mailer,
		WorkingFolderBase: workDir,
		BuddyHub:          bhub,
		BaseRouter:        tenantEngine,
		RunPortNoBind:     true,
		OnStart: func() {
			s.markReady(nil)
			log.Printf("tenant %s ready for in-process HTTP", s.Name)
		},
	})

	if err := seedTenantApp(happ, adminName, adminPassword, adminEmail); err != nil {
		return fmt.Errorf("tenant %s: seed: %w", s.Name, err)
	}

	s.App = happ
	s.Engine = tenantEngine

	go func() {
		if err := happ.Start(); err != nil {
			log.Printf("tenant %s start error: %v", s.Name, err)
			s.markReady(fmt.Errorf("tenant %s: start: %w", s.Name, err))
		}
	}()

	select {
	case <-s.readyCh:
	case <-time.After(30 * time.Second):
		return fmt.Errorf("tenant %s: timed out waiting for start", s.Name)
	}

	if s.startErr != nil {
		return s.startErr
	}

	if err := s.tursoDB.Push(ctx); err != nil {
		log.Printf("tenant %s: initial push: %v", s.Name, err)
	}

	return nil
}

func (s *SubApp) markReady(err error) {
	s.readyOnce.Do(func() {
		s.mu.Lock()
		s.startErr = err
		if err == nil {
			s.ready = true
		}
		s.mu.Unlock()
		close(s.readyCh)
	})
}

func (s *SubApp) WaitReady(timeout time.Duration) error {
	select {
	case <-s.readyCh:
		if s.startErr != nil {
			return s.startErr
		}
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("tenant %s: timed out waiting for start", s.Name)
	}
}

func (s *SubApp) Checkpoint(ctx context.Context) error {
	if s == nil || s.tursoDB == nil {
		return nil
	}
	return s.tursoDB.Checkpoint(ctx)
}

func (s *SubApp) IsReady() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ready
}

func (s *SubApp) Controller() (*actions.Controller, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.App == nil {
		return nil, fmt.Errorf("tenant %q is not loaded", s.Name)
	}
	ctrl, ok := s.App.Controller().(*actions.Controller)
	if !ok || ctrl == nil {
		return nil, fmt.Errorf("tenant %q controller unavailable", s.Name)
	}
	return ctrl, nil
}

func (s *SubApp) Unload() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.App != nil {
		if err := s.App.Close(); err != nil {
			log.Printf("tenant %s: app close: %v", s.Name, err)
		}
	}

	if s.tursoDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := s.tursoDB.Push(ctx); err != nil {
			log.Printf("tenant %s: unload push: %v", s.Name, err)
		}
		cancel()
	}
	s.App = nil
	s.Engine = nil
	s.ready = false
}

func seedTenantApp(happ *app.App, name, password, email string) error {
	ctrl := happ.Controller().(*actions.Controller)

	ugroups, err := ctrl.ListUserGroups()
	if err != nil {
		return fmt.Errorf("list user groups: %w", err)
	}
	if len(ugroups) == 0 {
		if err := ctrl.AddUserGroup("admin", "Admin group"); err != nil {
			return fmt.Errorf("add admin group: %w", err)
		}
		if err := ctrl.AddUserGroup("normal", "Normal group"); err != nil {
			return fmt.Errorf("add normal group: %w", err)
		}
	}

	users, err := ctrl.ListUsers(0, 1)
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}
	if len(users) > 0 {
		return nil
	}

	if name == "" {
		name = "admin"
	}
	if password == "" {
		password = "changeme_please_123"
	}
	if email == "" {
		email = "admin@localhost"
	}

	if _, err := ctrl.AddAdminUserDirect(name, password, email); err != nil {
		return fmt.Errorf("add admin user %q: %w", email, err)
	}

	return nil
}
