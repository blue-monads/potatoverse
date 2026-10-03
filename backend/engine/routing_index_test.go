package engine

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/blue-monads/potatoverse/backend/services/buddyhub"
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

func setupTestDB(t *testing.T) (datahub.Database, func()) {
	tmpDir, err := os.MkdirTemp("", "potatoverse_engine_test_*")
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

func TestRoutingIndex_AllPluginLoaderScript(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	eng := &Engine{
		db:     db,
		logger: slog.Default(),
	}
	sr := NewSpaceRouter(eng)
	eng.spaceRouter = sr

	now := time.Now()
	// 1. Create Package 1 (Host app: Accounting)
	res1, _ := db.Table("PackageInstalls").Insert(&dbmodels.InstalledPackage{
		Name:        "Super Accounting Package",
		Slug:        "super-acc",
		InstalledAt: &now,
	})
	pkg1Id := res1.ID().(int64)

	resV1, _ := db.Table("PackageVersion").Insert(&dbmodels.PackageVersion{
		InstallId: pkg1Id,
		Name:      "Super Accounting Package",
		Slug:      "super-acc",
		Version:   "1.0.0",
	})
	pver1Id := resV1.ID().(int64)
	db.GetPackageInstallOps().UpdateActiveInstallId(pkg1Id, pver1Id)

	// 2. Create Package 2 (Plugin: Inventory)
	res2, _ := db.Table("PackageInstalls").Insert(&dbmodels.InstalledPackage{
		Name:        "Super Inventory Package",
		Slug:        "super-inventry",
		InstalledAt: &now,
	})
	pkg2Id := res2.ID().(int64)

	resV2, _ := db.Table("PackageVersion").Insert(&dbmodels.PackageVersion{
		InstallId: pkg2Id,
		Name:      "Super Inventory Package",
		Slug:      "super-inventry",
		Version:   "1.0.0",
	})
	pver2Id := resV2.ID().(int64)
	db.GetPackageInstallOps().UpdateActiveInstallId(pkg2Id, pver2Id)

	// 3. Add loader.js file to Package 2 active version
	fakeLoaderContent := "console.log('inventory plugin loaded');"
	_, err := db.GetPackageFileOps().CreateFile(pver2Id, &datahub.CreateFileRequest{
		Name: "loader.js",
		Path: "",
	}, bytes.NewReader([]byte(fakeLoaderContent)))
	if err != nil {
		t.Fatalf("failed to create plugin loader file: %v", err)
	}

	// 4. Create Host Space
	space1Id, err := db.GetSpaceOps().AddSpace(&dbmodels.Space{
		InstalledId:  pkg1Id,
		NamespaceKey: "super-acc",
		SpaceType:    "App",
		RouteOptions: `{"router_type":"simple","serve_folder":"public"}`,
	})
	if err != nil {
		t.Fatalf("failed to add host space: %v", err)
	}

	// 5. Create Plugin Space
	pluginSpaceId, err := db.GetSpaceOps().AddSpace(&dbmodels.Space{
		InstalledId:  pkg2Id,
		NamespaceKey: "super-inventry:acc-integration",
		SpaceType:    "AppPlugin",
		LoaderScript: "loader.js",
		RouteOptions: `{"router_type":"simple","serve_folder":"public"}`,
	})
	if err != nil {
		t.Fatalf("failed to add plugin space: %v", err)
	}

	// 6. Connect plugin to host space via SpacePlugins
	_, err = db.GetSpaceOps().AddSpacePlugin(&dbmodels.SpacePlugin{
		SourceInstallID: pkg1Id,
		SourceSpaceID:   space1Id,
		TargetInstallID: pkg2Id,
		TargetSpaceID:   pluginSpaceId,
		ExtraMeta:       "{}",
	})
	if err != nil {
		t.Fatalf("failed to add space plugin: %v", err)
	}

	// 7. Build Index Item for host space
	hostSpace, _ := db.GetSpaceOps().GetSpace(space1Id)
	hostPkgVer, _ := db.GetPackageInstallOps().GetPackageVersion(pver1Id)

	indexItem, err := sr.buildIndexItem(hostSpace, hostPkgVer)
	if err != nil {
		t.Fatalf("buildIndexItem failed: %v", err)
	}

	if !strings.Contains(indexItem.allPluginLoaderScript, fakeLoaderContent) {
		t.Fatalf("expected allPluginLoaderScript to contain %q, got %q", fakeLoaderContent, indexItem.allPluginLoaderScript)
	}
	if !strings.Contains(indexItem.allPluginLoaderScript, "super-inventry:acc-integration") {
		t.Fatalf("expected allPluginLoaderScript to contain plugin namespace, got %q", indexItem.allPluginLoaderScript)
	}
}
