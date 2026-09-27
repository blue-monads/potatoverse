package actions

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/blue-monads/potatoverse/backend/services/buddyhub"
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/models"
)

func setupTestDB(t *testing.T) (datahub.Database, func()) {
	tmpDir, err := os.MkdirTemp("", "potatoverse_pkg_install_test_*")
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

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

func TestInstallArtifactSpace_AddsUserToSpaceUsers(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userId := int64(42)
	installedId := int64(100)
	artifact := &models.PotatoSpace{
		Namespace:    "test-pkg:test-space",
		ExecutorType: "core",
		ServerFile:   "index.js",
	}

	spaceId, err := installArtifactSpace(db, userId, installedId, artifact)
	if err != nil {
		t.Fatalf("installArtifactSpace failed: %v", err)
	}

	if spaceId == 0 {
		t.Fatalf("expected non-zero spaceId")
	}

	// Verify the space was created
	space, err := db.GetSpaceOps().GetSpace(spaceId)
	if err != nil {
		t.Fatalf("failed to get space: %v", err)
	}
	if space.OwnerID != userId {
		t.Errorf("expected space owner %d, got %d", userId, space.OwnerID)
	}

	// Verify user was added to space users
	spaceUsers, err := db.GetSpaceOps().ListSpaceUsers(spaceId)
	if err != nil {
		t.Fatalf("failed to list space users: %v", err)
	}

	if len(spaceUsers) != 1 {
		t.Fatalf("expected 1 space user, got %d", len(spaceUsers))
	}

	su := spaceUsers[0]
	if su.UserID != userId {
		t.Errorf("expected user_id %d, got %d", userId, su.UserID)
	}
	if su.SpaceID != spaceId {
		t.Errorf("expected space_id %d, got %d", spaceId, su.SpaceID)
	}
	if su.InstallID != installedId {
		t.Errorf("expected install_id %d, got %d", installedId, su.InstallID)
	}

	// Verify querying space users by installId also returns the user
	queriedUsers, err := db.GetSpaceOps().QuerySpaceUsers(installedId, map[any]any{
		"space_id": spaceId,
	})
	if err != nil {
		t.Fatalf("failed to query space users: %v", err)
	}
	if len(queriedUsers) != 1 {
		t.Fatalf("expected 1 queried space user, got %d", len(queriedUsers))
	}
	if queriedUsers[0].UserID != userId {
		t.Errorf("expected user_id %d, got %d", userId, queriedUsers[0].UserID)
	}
}

func TestInstallArtifactSpace_ZeroUserId_DoesNotAddSpaceUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userId := int64(0)
	installedId := int64(100)
	artifact := &models.PotatoSpace{
		Namespace:    "test-pkg:zero-user-space",
		ExecutorType: "core",
	}

	spaceId, err := installArtifactSpace(db, userId, installedId, artifact)
	if err != nil {
		t.Fatalf("installArtifactSpace failed: %v", err)
	}

	spaceUsers, err := db.GetSpaceOps().ListSpaceUsers(spaceId)
	if err != nil {
		t.Fatalf("failed to list space users: %v", err)
	}

	if len(spaceUsers) != 0 {
		t.Errorf("expected 0 space users when userId is 0, got %d", len(spaceUsers))
	}
}
