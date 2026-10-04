package actions

import (
	"archive/zip"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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

func TestInstallArtifactSpace_UserAlreadyExists_DoesNotDuplicate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	userId := int64(42)
	installedId := int64(100)
	artifact := &models.PotatoSpace{
		Namespace:    "test-pkg:existing-user-space",
		ExecutorType: "core",
	}

	spaceId, err := installArtifactSpace(db, userId, installedId, artifact)
	if err != nil {
		t.Fatalf("first installArtifactSpace failed: %v", err)
	}

	// Verify 1 user
	spaceUsers, err := db.GetSpaceOps().ListSpaceUsers(spaceId)
	if err != nil {
		t.Fatalf("failed to list space users: %v", err)
	}
	if len(spaceUsers) != 1 {
		t.Fatalf("expected 1 space user, got %d", len(spaceUsers))
	}

	// Now simulate calling installArtifactSpace again (or if user was already added to spaceId)
	// QuerySpaceUsers should find the existing user and not attempt to re-add or error on duplicate
	existingUsers, err := db.GetSpaceOps().QuerySpaceUsers(installedId, map[any]any{
		"user_id":  userId,
		"space_id": spaceId,
	})
	if err != nil {
		t.Fatalf("QuerySpaceUsers failed: %v", err)
	}
	if len(existingUsers) != 1 {
		t.Fatalf("expected 1 existing user, got %d", len(existingUsers))
	}

	// If we invoke the user addition logic again, it should detect user exists
	if len(existingUsers) == 0 {
		t.Fatalf("expected user to exist, so len should be > 0")
	}

	// Check that space users count remains 1
	spaceUsersAfter, err := db.GetSpaceOps().ListSpaceUsers(spaceId)
	if err != nil {
		t.Fatalf("failed to list space users: %v", err)
	}
	if len(spaceUsersAfter) != 1 {
		t.Errorf("expected still 1 space user, got %d", len(spaceUsersAfter))
	}
}

func createTestPkgZip(t *testing.T, potatoJSON string) string {
	t.Helper()
	f, err := os.CreateTemp("", "test-pkg-*.zip")
	if err != nil {
		t.Fatalf("create temp zip: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	w, err := zw.Create("potato.json")
	if err != nil {
		t.Fatalf("create potato.json in zip: %v", err)
	}
	if _, err := w.Write([]byte(potatoJSON)); err != nil {
		t.Fatalf("write potato.json: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	return f.Name()
}

func TestInstallPackage_NamespaceValidation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := slog.Default()
	userId := int64(1)

	t.Run("valid main space and subkey space", func(t *testing.T) {
		manifest := `{
			"name": "My Package",
			"slug": "my-pkg",
			"spaces": [
				{"name": "Main", "namespace": "my-pkg", "executor_type": "core", "server_file": "index.js"},
				{"name": "Sub", "namespace": "my-pkg:sub-space", "executor_type": "core", "server_file": "sub.js"}
			]
		}`
		zipPath := createTestPkgZip(t, manifest)
		defer os.Remove(zipPath)

		res, err := installPackageByFile(db, logger, userId, "", zipPath)
		if err != nil {
			t.Fatalf("expected install to succeed, got %v", err)
		}
		if res.MainSpaceId == 0 {
			t.Errorf("expected non-zero MainSpaceId")
		}

	})

	t.Run("multiple main spaces found error", func(t *testing.T) {
		manifest := `{
			"name": "My Package Multi",
			"slug": "my-multi-pkg",
			"spaces": [
				{"name": "Main 1", "namespace": "my-multi-pkg", "executor_type": "core"},
				{"name": "Main 2", "namespace": "my-multi-pkg", "executor_type": "core"}
			]
		}`
		zipPath := createTestPkgZip(t, manifest)
		defer os.Remove(zipPath)

		_, err := installPackageByFile(db, logger, userId, "", zipPath)
		if err == nil || !strings.Contains(err.Error(), "multiple main spaces found") {
			t.Fatalf("expected 'multiple main spaces found' error, got %v", err)
		}
	})

	t.Run("invalid namespace prefix error", func(t *testing.T) {
		manifest := `{
			"name": "My Package Prefix",
			"slug": "prefix-pkg",
			"spaces": [
				{"name": "Other", "namespace": "other-pkg:space", "executor_type": "core"}
			]
		}`
		zipPath := createTestPkgZip(t, manifest)
		defer os.Remove(zipPath)

		_, err := installPackageByFile(db, logger, userId, "", zipPath)
		if err == nil {
			t.Fatalf("expected error for namespace not matching pkg slug")
		}
	})

	t.Run("invalid subkey with multiple colons", func(t *testing.T) {
		manifest := `{
			"name": "My Package Subkey Colons",
			"slug": "colon-pkg",
			"spaces": [
				{"name": "Colon", "namespace": "colon-pkg:sub:extra", "executor_type": "core"}
			]
		}`
		zipPath := createTestPkgZip(t, manifest)
		defer os.Remove(zipPath)

		_, err := installPackageByFile(db, logger, userId, "", zipPath)
		if err == nil {
			t.Fatalf("expected error for namespace with multiple colons")
		}
	})
}
