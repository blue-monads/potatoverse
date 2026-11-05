package server

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/blue-monads/potatoverse/backend/app/actions"
	rtbuddy "github.com/blue-monads/potatoverse/backend/app/server/rt_buddy"
	"github.com/blue-monads/potatoverse/backend/engine"
	"github.com/blue-monads/potatoverse/backend/services/buddyhub"
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
)

func strPtr(s string) *string {
	return &s
}

func setupTestRoutingServer(t *testing.T) (*Server, datahub.Database, *signer.Signer, func()) {
	gin.SetMode(gin.TestMode)

	tmpDir, err := os.MkdirTemp("", "potatoverse_routing_test_*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}

	logger := slog.Default()
	dbFile := filepath.Join(tmpDir, "test.sqlite")
	db, err := database.NewDB(dbFile, logger)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("new db: %v", err)
	}

	opts := &xtypes.AppOptions{
		Name:         "PotatoTest",
		Port:         8080,
		WorkingDir:   tmpDir,
		MasterSecret: "super_secret_master_key_123456",
	}

	bhub := buddyhub.NewBuddyHub(opts, logger)
	if err := db.Init(bhub); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("init db: %v", err)
	}

	sig := signer.New([]byte("super_secret_master_key_123456"))

	eng := engine.NewEngine(engine.EngineOption{
		DB:            db,
		Logger:        logger,
		WorkingFolder: tmpDir,
	})

	ctrl := actions.New(actions.Option{
		Database: db,
		Logger:   logger,
		Signer:   sig,
		Engine:   eng,
	})

	router := gin.New()

	s := &Server{
		router:      router,
		ctrl:        ctrl,
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
	s.bindRoutes()

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return s, db, sig, cleanup
}

func TestRoutingAPI_AdminOnly(t *testing.T) {
	srv, db, sig, cleanup := setupTestRoutingServer(t)
	defer cleanup()

	// 1. Create admin user (id=1) and normal user (id=2)
	adminUser := &dbmodels.User{
		Name:     "Admin",
		Username: strPtr("admin"),
		Email:    "admin@example.com",
		Password: "secret",
		Ugroup:   "admin",
	}
	adminId, err := db.GetUserOps().AddUser(adminUser)
	if err != nil {
		t.Fatalf("add admin user: %v", err)
	}

	normalUser := &dbmodels.User{
		Name:     "Normal",
		Username: strPtr("normal"),
		Email:    "normal@example.com",
		Password: "secret",
		Ugroup:   "normal",
	}
	normalId, err := db.GetUserOps().AddUser(normalUser)
	if err != nil {
		t.Fatalf("add normal user: %v", err)
	}

	normalToken, _ := sig.SignAccess(&signer.AccessClaim{UserId: normalId})
	adminToken, _ := sig.SignAccess(&signer.AccessClaim{UserId: adminId})

	// 2. Normal user tries to GET /zz/api/core/admin/routing/root -> 403 Forbidden
	req, _ := http.NewRequest("GET", "/zz/api/core/admin/routing/root", nil)
	req.Header.Set("Authorization", normalToken)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 Forbidden for normal user, got %d: %s", w.Code, w.Body.String())
	}

	// 3. Admin user tries GET -> 200 OK
	req, _ = http.NewRequest("GET", "/zz/api/core/admin/routing/root", nil)
	req.Header.Set("Authorization", adminToken)
	w = httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK for admin user, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRoutingAPI_MainSpaceValidationAndReload(t *testing.T) {
	srv, db, sig, cleanup := setupTestRoutingServer(t)
	defer cleanup()

	adminUser := &dbmodels.User{
		Name:     "Admin",
		Username: strPtr("admin"),
		Email:    "admin@example.com",
		Password: "secret",
		Ugroup:   "admin",
	}
	adminId, err := db.GetUserOps().AddUser(adminUser)
	if err != nil {
		t.Fatalf("add admin user: %v", err)
	}
	adminToken, _ := sig.SignAccess(&signer.AccessClaim{UserId: adminId})

	// 1. Create a plugin space (type AppPlugin)
	pluginSpace := &dbmodels.Space{
		InstalledId:  1,
		NamespaceKey: "my-plugin",
		SpaceType:    "AppPlugin",
		OwnerID:      adminId,
	}
	pluginSpaceId, err := db.GetSpaceOps().AddSpace(pluginSpace)
	if err != nil {
		t.Fatalf("add plugin space: %v", err)
	}

	// Try to configure root routing with space of type "AppPlugin" -> Should fail validation
	body, _ := json.Marshal(map[string]any{
		"routes": map[string]any{
			"example.com": map[string]any{"space_id": pluginSpaceId},
		},
	})
	req, _ := http.NewRequest("POST", "/zz/api/core/admin/routing/root", bytes.NewReader(body))
	req.Header.Set("Authorization", adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected error when routing AppPlugin space, but got 200 OK: %s", w.Body.String())
	}

	// 2. Create a subkey space (contains colon)
	subkeySpace := &dbmodels.Space{
		InstalledId:  1,
		NamespaceKey: "my-app:subkey",
		SpaceType:    "App",
		OwnerID:      adminId,
	}
	subkeySpaceId, err := db.GetSpaceOps().AddSpace(subkeySpace)
	if err != nil {
		t.Fatalf("add subkey space: %v", err)
	}

	// Try to configure root routing with subkey space -> Should fail validation
	body, _ = json.Marshal(map[string]any{
		"routes": map[string]any{
			"example.com": map[string]any{"space_id": subkeySpaceId},
		},
	})
	req, _ = http.NewRequest("POST", "/zz/api/core/admin/routing/root", bytes.NewReader(body))
	req.Header.Set("Authorization", adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected error when routing subkey space, but got 200 OK: %s", w.Body.String())
	}

	// 3. Create a valid Main Space (App, no colon in namespace)
	mainSpace := &dbmodels.Space{
		InstalledId:  1,
		NamespaceKey: "my-app",
		SpaceType:    "App",
		OwnerID:      adminId,
	}
	mainSpaceId, err := db.GetSpaceOps().AddSpace(mainSpace)
	if err != nil {
		t.Fatalf("add main space: %v", err)
	}

	// Now configure root routing with the Main Space -> Should succeed
	rootBody, _ := json.Marshal(map[string]any{
		"routes": map[string]any{
			"example.com": map[string]any{"space_id": mainSpaceId},
		},
	})
	req, _ = http.NewRequest("POST", "/zz/api/core/admin/routing/root", bytes.NewReader(rootBody))
	req.Header.Set("Authorization", adminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected root routing save to succeed, got %d: %s", w.Code, w.Body.String())
	}

	// Verify engine router matches domain immediately
	matchedSpaceId, found := srv.engine.GetRootRouter().MatchDomain("example.com")
	if !found || matchedSpaceId != mainSpaceId {
		t.Fatalf("expected engine router to match domain 'example.com' to space %d, got %d (found=%v)", mainSpaceId, matchedSpaceId, found)
	}

	// Test reload endpoint
	req, _ = http.NewRequest("POST", "/zz/api/core/admin/routing/root/reload", nil)
	req.Header.Set("Authorization", adminToken)
	w = httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected reload endpoint to succeed, got %d: %s", w.Code, w.Body.String())
	}
}
