package actions

import (
	"log/slog"
	"testing"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/xtypes/models"
)

func TestSpacePlugin_CRUDAndFiltering(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctrl := New(Option{
		Database: db,
		Logger:   slog.Default(),
	})

	userId := int64(1)
	now := time.Now()

	// 1. Create package 1 (Accounting) and package 2 (Inventory)
	res1, err := db.Table("PackageInstalls").Insert(&dbmodels.InstalledPackage{
		Name:        "Super Accounting Package",
		Slug:        "super-acc",
		InstalledBy: userId,
		InstalledAt: &now,
	})
	if err != nil {
		t.Fatalf("failed to add pkg1: %v", err)
	}
	pkg1Id := res1.ID().(int64)

	resV1, err := db.Table("PackageVersion").Insert(&dbmodels.PackageVersion{
		InstallId:  pkg1Id,
		Name:       "Super Accounting Package",
		Slug:       "super-acc",
		Version:    "1.0.0",
		AuthorName: "Alice",
	})
	if err != nil {
		t.Fatalf("failed to add pver1: %v", err)
	}
	pver1Id := resV1.ID().(int64)
	db.GetPackageInstallOps().UpdateActiveInstallId(pkg1Id, pver1Id)

	res2, err := db.Table("PackageInstalls").Insert(&dbmodels.InstalledPackage{
		Name:        "Super Inventory Package",
		Slug:        "super-inventry",
		InstalledBy: userId,
		InstalledAt: &now,
	})
	if err != nil {
		t.Fatalf("failed to add pkg2: %v", err)
	}
	pkg2Id := res2.ID().(int64)

	resV2, err := db.Table("PackageVersion").Insert(&dbmodels.PackageVersion{
		InstallId:  pkg2Id,
		Name:       "Super Inventory Package",
		Slug:       "super-inventry",
		Version:    "1.0.0",
		AuthorName: "Bob",
	})
	if err != nil {
		t.Fatalf("failed to add pver2: %v", err)
	}
	pver2Id := resV2.ID().(int64)
	db.GetPackageInstallOps().UpdateActiveInstallId(pkg2Id, pver2Id)

	// 2. Install App space in pkg1 and AppPlugin space in pkg2
	space1Id, err := installArtifactSpace(db, userId, pkg1Id, &models.PotatoSpace{
		Namespace:    "super-acc",
		Type:         "App",
		ExecutorType: "luaz",
		ServerFile:   "server.lua",
	})
	if err != nil {
		t.Fatalf("failed to install space1: %v", err)
	}

	pluginSpaceId, err := installArtifactSpace(db, userId, pkg2Id, &models.PotatoSpace{
		Namespace:    "super-inventry:acc-integration",
		Type:         "AppPlugin",
		LoaderScript: "loader.js",
		ExecutorType: "luaz",
		ServerFile:   "server.lua",
	})
	if err != nil {
		t.Fatalf("failed to install pluginSpace: %v", err)
	}

	// 3. Test ListInstalledSpaces includes AppPlugin (filtering is handled in frontend)
	installedSpaces, err := ctrl.ListInstalledSpaces(userId)
	if err != nil {
		t.Fatalf("ListInstalledSpaces failed: %v", err)
	}

	var foundPluginSpace bool
	for _, s := range installedSpaces.Spaces {
		if s.ID == pluginSpaceId && s.SpaceType == "AppPlugin" {
			foundPluginSpace = true
			break
		}
	}
	if !foundPluginSpace {
		t.Fatalf("ListInstalledSpaces should include AppPlugin space %d", pluginSpaceId)
	}

	var foundSpace1 bool
	for _, s := range installedSpaces.Spaces {
		if s.ID == space1Id {
			foundSpace1 = true
			break
		}
	}
	if !foundSpace1 {
		t.Fatalf("ListInstalledSpaces should include App space %d", space1Id)
	}

	// 4. Test ListAvailableAppPlugins for space1
	avail, err := ctrl.ListAvailableAppPlugins(pkg1Id, space1Id)
	if err != nil {
		t.Fatalf("ListAvailableAppPlugins failed: %v", err)
	}
	if len(avail) != 1 {
		t.Fatalf("expected 1 available plugin, got %d", len(avail))
	}
	if avail[0].SpaceID != pluginSpaceId {
		t.Fatalf("expected available plugin space_id %d, got %d", pluginSpaceId, avail[0].SpaceID)
	}
	if avail[0].IsAlreadyPlugged {
		t.Fatalf("plugin should not be plugged yet")
	}
	if avail[0].LoaderScript != "loader.js" {
		t.Fatalf("expected loader_script loader.js, got %s", avail[0].LoaderScript)
	}

	// 5. Plug pluginSpace into space1
	pluginRecord, err := ctrl.CreateSpacePlugin(pkg1Id, space1Id, map[string]any{
		"target_install_id": pkg2Id,
		"target_space_id":   pluginSpaceId,
		"extrameta":         `{"enabled":true}`,
	})
	if err != nil {
		t.Fatalf("CreateSpacePlugin failed: %v", err)
	}
	if pluginRecord == nil || pluginRecord.ID == 0 {
		t.Fatalf("expected valid plugin record, got %v", pluginRecord)
	}

	// 6. Test ListSpacePlugins returns enriched plugin
	pluggedList, err := ctrl.ListSpacePlugins(pkg1Id, space1Id)
	if err != nil {
		t.Fatalf("ListSpacePlugins failed: %v", err)
	}
	if len(pluggedList) != 1 {
		t.Fatalf("expected 1 plugged plugin, got %d", len(pluggedList))
	}
	if pluggedList[0].TargetNamespaceKey != "super-inventry:acc-integration" {
		t.Fatalf("expected TargetNamespaceKey super-inventry:acc-integration, got %s", pluggedList[0].TargetNamespaceKey)
	}
	if pluggedList[0].TargetPackageName != "Super Inventory Package" {
		t.Fatalf("expected TargetPackageName Super Inventory Package, got %s", pluggedList[0].TargetPackageName)
	}
	if pluggedList[0].TargetLoaderScript != "loader.js" {
		t.Fatalf("expected TargetLoaderScript loader.js, got %s", pluggedList[0].TargetLoaderScript)
	}

	// 7. Verify ListAvailableAppPlugins now reports IsAlreadyPlugged = true
	availAfter, err := ctrl.ListAvailableAppPlugins(pkg1Id, space1Id)
	if err != nil {
		t.Fatalf("ListAvailableAppPlugins after plugging failed: %v", err)
	}
	if len(availAfter) != 1 || !availAfter[0].IsAlreadyPlugged {
		t.Fatalf("expected plugin to be marked as already plugged")
	}

	// 8. Update plugin extrameta
	updated, err := ctrl.UpdateSpacePluginByID(pkg1Id, pluginRecord.ID, map[string]any{
		"extrameta": `{"enabled":false}`,
	})
	if err != nil {
		t.Fatalf("UpdateSpacePluginByID failed: %v", err)
	}
	if updated.ExtraMeta != `{"enabled":false}` {
		t.Fatalf("expected updated extrameta, got %s", updated.ExtraMeta)
	}

	// 9. Delete plugin
	err = ctrl.DeleteSpacePluginByID(pkg1Id, pluginRecord.ID)
	if err != nil {
		t.Fatalf("DeleteSpacePluginByID failed: %v", err)
	}

	pluggedAfterDelete, err := ctrl.ListSpacePlugins(pkg1Id, space1Id)
	if err != nil {
		t.Fatalf("ListSpacePlugins after delete failed: %v", err)
	}
	if len(pluggedAfterDelete) != 0 {
		t.Fatalf("expected 0 plugged plugins after delete, got %d", len(pluggedAfterDelete))
	}
}
