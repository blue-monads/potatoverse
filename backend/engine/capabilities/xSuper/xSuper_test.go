package xsuper

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/blue-monads/potatoverse/backend/app/actions"
	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/services/buddyhub"
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
)

type mockApp struct {
	db      datahub.Database
	sig     *signer.Signer
	logger  *slog.Logger
	ctrl    *actions.Controller
	appOpts *xtypes.AppOptions
}

func (m *mockApp) ExecId() string             { return "test-exec" }
func (m *mockApp) Init() error                { return nil }
func (m *mockApp) Start() error               { return nil }
func (m *mockApp) Close() error               { return nil }
func (m *mockApp) Database() datahub.Database { return m.db }
func (m *mockApp) Signer() *signer.Signer     { return m.sig }
func (m *mockApp) Logger() *slog.Logger       { return m.logger }
func (m *mockApp) Controller() any            { return m.ctrl }
func (m *mockApp) Engine() any                { return nil }
func (m *mockApp) Config() any                { return m.appOpts }
func (m *mockApp) Sockd() any                 { return nil }
func (m *mockApp) CoreHub() any               { return nil }

type mockHandle struct {
	model *dbmodels.SpaceCapability
}

var _ xcapability.XCapabilityHandle = (*mockHandle)(nil)

func (h *mockHandle) GetModel() *dbmodels.SpaceCapability {
	return h.model
}
func (h *mockHandle) GetSpaceId() (int64, error) {
	return h.model.SpaceID, nil
}
func (h *mockHandle) ParseCapToken(token string) (*signer.CapabilityClaim, error) {
	return &signer.CapabilityClaim{}, nil
}
func (h *mockHandle) ValidateCapToken(token string) (*signer.CapabilityClaim, error) {
	return &signer.CapabilityClaim{}, nil
}
func (h *mockHandle) GetOptions(target any) error {
	return nil
}
func (h *mockHandle) GetOptionsAsLazyData() lazydata.LazyData {
	return nil
}

func setupTestApp(t *testing.T) (*mockApp, func()) {
	tmpDir, err := os.MkdirTemp("", "potatoverse_xsuper_test_*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	opts := &xtypes.AppOptions{
		Name:         "PotatoTest",
		Port:         8080,
		Hosts:        []xtypes.Host{{Name: "example.com"}},
		MasterSecret: "super_secret_master_key_123456",
		WorkingDir:   tmpDir,
	}

	logger := slog.Default()
	bhub := buddyhub.NewBuddyHub(opts, logger)
	dbFile := filepath.Join(tmpDir, "test.sqlite")
	db, err := database.NewDB(dbFile, logger)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("new db: %v", err)
	}
	if err := db.Init(bhub); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("init db: %v", err)
	}

	sig := signer.New([]byte(opts.MasterSecret))
	ctrl := actions.New(actions.Option{
		Database: db,
		Logger:   logger,
		Signer:   sig,
		AppOpts:  opts,
	})

	if err := ctrl.AddUserGroup("admin", "Admin group"); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("add admin group: %v", err)
	}
	if _, err := ctrl.AddAdminUserDirect("adminuser", "adminpass123", "admin@example.com"); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("add admin user: %v", err)
	}

	app := &mockApp{
		db:      db,
		sig:     sig,
		logger:  logger,
		ctrl:    ctrl,
		appOpts: opts,
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return app, cleanup
}

func TestSuperCapability(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	factories, err := registry.GetCapabilityBuilderFactories()
	if err != nil {
		t.Fatalf("failed to get capability factories: %v", err)
	}
	factory, ok := factories["xSuper"]
	if !ok {
		t.Fatalf("xSuper capability not registered")
	}

	builder, err := factory.Builder(app)
	if err != nil {
		t.Fatalf("failed to build capability builder: %v", err)
	}

	handle := &mockHandle{
		model: &dbmodels.SpaceCapability{
			ID:        1,
			SpaceID:   10,
			InstallID: 20,
		},
	}

	cap, err := builder.Build(handle)
	if err != nil {
		t.Fatalf("failed to build capability: %v", err)
	}

	t.Run("list actions", func(t *testing.T) {
		actions, err := cap.ListActions()
		if err != nil {
			t.Fatalf("list actions failed: %v", err)
		}
		if len(actions) != 2 {
			t.Fatalf("expected 2 actions, got %v", actions)
		}
	})

	t.Run("get_instance_info", func(t *testing.T) {
		res, err := cap.Execute("get_instance_info", nil)
		if err != nil {
			t.Fatalf("get_instance_info failed: %v", err)
		}
		data, ok := res.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", res)
		}
		if data["port"] != 8080 {
			t.Fatalf("expected port 8080, got %v", data["port"])
		}
		if data["host"] != "example.com" {
			t.Fatalf("expected host example.com, got %v", data["host"])
		}
	})

	t.Run("get_admin_token", func(t *testing.T) {
		res, err := cap.Execute("get_admin_token", nil)
		if err != nil {
			t.Fatalf("get_admin_token failed: %v", err)
		}
		data, ok := res.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", res)
		}
		token, ok := data["token"].(string)
		if !ok || token == "" {
			t.Fatalf("expected non-empty token, got %v", data["token"])
		}
		user, ok := data["user"].(*dbmodels.User)
		if !ok || user == nil {
			t.Fatalf("expected dbmodels.User, got %T", data["user"])
		}
		if user.Email != "admin@example.com" {
			t.Fatalf("expected admin@example.com, got %s", user.Email)
		}
	})
}
