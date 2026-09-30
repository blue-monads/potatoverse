package actions

import (
	"log/slog"
	"testing"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/signer"
)

func TestSpaceUserGroup_CRUD(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctrl := New(Option{
		Database: db,
		Logger:   slog.Default(),
		Signer:   signer.New([]byte("test_secret")),
	})

	err := db.GetUserOps().AddUserGroup("finance", "Finance Department")
	if err != nil {
		t.Fatalf("failed to add user group: %v", err)
	}
	ug, err := db.GetUserOps().GetUserGroup("finance")
	if err != nil {
		t.Fatalf("failed to get user group: %v", err)
	}

	installId := int64(10)
	spaceId := int64(20)

	// Create SpaceUserGroup
	sug, err := ctrl.CreateSpaceUserGroup(installId, map[string]any{
		"group_id": ug.ID,
		"space_id": spaceId,
		"scope":    "viewer",
	})
	if err != nil {
		t.Fatalf("CreateSpaceUserGroup failed: %v", err)
	}
	if sug.ID == 0 {
		t.Fatalf("expected non-zero ID")
	}
	if sug.GroupID != ug.ID || sug.SpaceID != spaceId || sug.Scope != "viewer" {
		t.Fatalf("unexpected space user group fields: %+v", sug)
	}

	// Get by ID
	fetched, err := ctrl.GetSpaceUserGroupByID(installId, sug.ID)
	if err != nil {
		t.Fatalf("GetSpaceUserGroupByID failed: %v", err)
	}
	if fetched.ID != sug.ID || fetched.Scope != "viewer" {
		t.Fatalf("fetched does not match: %+v", fetched)
	}

	// Query
	groups, err := ctrl.QuerySpaceUserGroups(installId, map[any]any{"group_id": ug.ID})
	if err != nil {
		t.Fatalf("QuerySpaceUserGroups failed: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}

	// Update
	updated, err := ctrl.UpdateSpaceUserGroupByID(installId, sug.ID, map[string]any{
		"scope": "editor",
	})
	if err != nil {
		t.Fatalf("UpdateSpaceUserGroupByID failed: %v", err)
	}
	if updated.Scope != "editor" {
		t.Fatalf("expected scope 'editor', got %s", updated.Scope)
	}

	// Delete
	err = ctrl.DeleteSpaceUserGroupByID(installId, sug.ID)
	if err != nil {
		t.Fatalf("DeleteSpaceUserGroupByID failed: %v", err)
	}

	groupsAfter, err := ctrl.QuerySpaceUserGroups(installId, map[any]any{})
	if err != nil {
		t.Fatalf("QuerySpaceUserGroups after delete failed: %v", err)
	}
	if len(groupsAfter) != 0 {
		t.Fatalf("expected 0 groups after delete, got %d", len(groupsAfter))
	}
}

func TestAuthorizeSpace_GroupAccess(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctrl := New(Option{
		Database: db,
		Logger:   slog.Default(),
		Signer:   signer.New([]byte("test_secret")),
	})

	// Create user group "analysts"
	err := db.GetUserOps().AddUserGroup("analysts", "Data Analysts")
	if err != nil {
		t.Fatalf("failed to add user group: %v", err)
	}
	ug, err := db.GetUserOps().GetUserGroup("analysts")
	if err != nil {
		t.Fatalf("failed to get user group: %v", err)
	}

	// Create user "bob" belonging to "analysts"
	bobId, err := db.GetUserOps().AddUser(&dbmodels.User{
		Name:     "Bob",
		Email:    "bob@example.com",
		Utype:    "user",
		Ugroup:   "analysts",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}

	ownerId := int64(999)
	installId := int64(100)

	// Create Space owned by ownerId
	spaceId, err := db.GetSpaceOps().AddSpace(&dbmodels.Space{
		InstalledId:  installId,
		NamespaceKey: "analytics",
		SpaceType:    "App",
		OwnerID:      ownerId,
	})
	if err != nil {
		t.Fatalf("failed to add space: %v", err)
	}

	// 1. Bob tries to authorize before group is added -> should fail
	_, err = ctrl.AuthorizeSpace(bobId, SpaceAuth{SpaceId: spaceId})
	if err != ErrUserNotAllowed {
		t.Fatalf("expected ErrUserNotAllowed, got %v", err)
	}

	// 2. Add group to Space (space-level)
	sug, err := ctrl.CreateSpaceUserGroup(installId, map[string]any{
		"group_id": ug.ID,
		"space_id": spaceId,
	})
	if err != nil {
		t.Fatalf("failed to add group to space: %v", err)
	}

	// Bob tries to authorize again -> should succeed
	token, err := ctrl.AuthorizeSpace(bobId, SpaceAuth{SpaceId: spaceId})
	if err != nil {
		t.Fatalf("expected AuthorizeSpace to succeed, got %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}

	// 3. Remove space-level entry, add package-level entry (space_id = 0)
	err = ctrl.DeleteSpaceUserGroupByID(installId, sug.ID)
	if err != nil {
		t.Fatalf("failed to delete space user group: %v", err)
	}

	_, err = ctrl.CreateSpaceUserGroup(installId, map[string]any{
		"group_id": ug.ID,
		"space_id": 0,
	})
	if err != nil {
		t.Fatalf("failed to add package-level group: %v", err)
	}

	// Bob tries to authorize via package-level group -> should succeed
	token2, err := ctrl.AuthorizeSpace(bobId, SpaceAuth{SpaceId: spaceId})
	if err != nil {
		t.Fatalf("expected AuthorizeSpace to succeed via package-level group, got %v", err)
	}
	if token2 == "" {
		t.Fatalf("expected non-empty token")
	}
}

func TestListInstalledSpaces_GroupAccess(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctrl := New(Option{
		Database: db,
		Logger:   slog.Default(),
		Signer:   signer.New([]byte("test_secret")),
	})

	now := time.Now()

	// Create user group "operators"
	err := db.GetUserOps().AddUserGroup("operators", "Operations")
	if err != nil {
		t.Fatalf("failed to add user group: %v", err)
	}
	ug, err := db.GetUserOps().GetUserGroup("operators")
	if err != nil {
		t.Fatalf("failed to get user group: %v", err)
	}

	// Create user "charlie" belonging to "operators"
	charlieId, err := db.GetUserOps().AddUser(&dbmodels.User{
		Name:     "Charlie",
		Email:    "charlie@example.com",
		Utype:    "user",
		Ugroup:   "operators",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("failed to add user: %v", err)
	}

	ownerId := int64(888)
	// Create PackageInstall
	resPkg, err := db.Table("PackageInstalls").Insert(&dbmodels.InstalledPackage{
		Name:        "Ops Portal",
		Slug:        "ops-portal",
		InstalledBy: ownerId,
		InstalledAt: &now,
	})
	if err != nil {
		t.Fatalf("failed to insert package install: %v", err)
	}
	pkgId := resPkg.ID().(int64)

	resVer, err := db.Table("PackageVersion").Insert(&dbmodels.PackageVersion{
		InstallId: pkgId,
		Name:      "Ops Portal",
		Slug:      "ops-portal",
		Version:   "1.0.0",
	})
	if err != nil {
		t.Fatalf("failed to insert package version: %v", err)
	}
	verId := resVer.ID().(int64)
	db.GetPackageInstallOps().UpdateActiveInstallId(pkgId, verId)

	// Create Space
	spaceId, err := db.GetSpaceOps().AddSpace(&dbmodels.Space{
		InstalledId:  pkgId,
		NamespaceKey: "ops",
		SpaceType:    "App",
		OwnerID:      ownerId,
	})
	if err != nil {
		t.Fatalf("failed to add space: %v", err)
	}

	// 1. Charlie lists installed spaces before group is added -> 0 spaces
	resInitial, err := ctrl.ListInstalledSpaces(charlieId)
	if err != nil {
		t.Fatalf("ListInstalledSpaces failed: %v", err)
	}
	if len(resInitial.Spaces) != 0 {
		t.Fatalf("expected 0 spaces, got %d", len(resInitial.Spaces))
	}

	// 2. Add group to package level
	_, err = ctrl.CreateSpaceUserGroup(pkgId, map[string]any{
		"group_id": ug.ID,
		"space_id": 0,
	})
	if err != nil {
		t.Fatalf("failed to add group to package: %v", err)
	}

	// 3. Charlie lists installed spaces -> should include the space
	resAfter, err := ctrl.ListInstalledSpaces(charlieId)
	if err != nil {
		t.Fatalf("ListInstalledSpaces failed: %v", err)
	}
	if len(resAfter.Spaces) != 1 {
		t.Fatalf("expected 1 space, got %d", len(resAfter.Spaces))
	}
	if resAfter.Spaces[0].ID != spaceId {
		t.Fatalf("expected space ID %d, got %d", spaceId, resAfter.Spaces[0].ID)
	}
}
