package migrator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

type MigratorCapability struct {
	folder  string
	builder *MigratorBuilder

	installId    int64
	installPvId  int64
	spaceId      int64
	capabilityId int64
	db           datahub.DBLowOps
}

func (m *MigratorCapability) Handle(ctx *gin.Context) {}

func (m *MigratorCapability) ListActions() ([]string, error) {
	return []string{"run_migrations", "list_migrations"}, nil
}

func (m *MigratorCapability) Execute(name string, params lazydata.LazyData) (any, error) {
	switch name {
	case "run_migrations":

		folder := params.GetFieldAsString("folder")

		if err := m.performMigration(folder); err != nil {
			return nil, fmt.Errorf("migration failed: %w", err)
		}
		return map[string]any{"status": "success", "message": "migrations completed"}, nil

	case "list_migrations":
		executed, err := m.getExecutedMigrations()
		if err != nil {
			return nil, err
		}

		folder := params.GetFieldAsString("folder")

		if folder == "" {
			folder = m.folder
		}

		pkgFileOps := m.builder.app.Database().GetPackageFileOps()
		files, err := pkgFileOps.ListFiles(m.installId, folder)
		if err != nil {
			return nil, err
		}

		migrations := []map[string]any{}
		for _, file := range files {
			if !file.IsFolder && strings.HasSuffix(strings.ToLower(file.Name), ".sql") {

				migrationKey := getMigrationKey(folder, file)
				migrations = append(migrations, map[string]any{
					"file_name":     file.Name,
					"path":          file.Path,
					"executed":      executed[migrationKey],
					"migration_key": migrationKey,
				})
			}
		}

		return map[string]any{"migrations": migrations}, nil

	default:
		return nil, fmt.Errorf("invalid action: %s", name)
	}
}

func (m *MigratorCapability) Reload(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {

	return &MigratorCapability{
		folder:       m.folder,
		builder:      m.builder,
		installPvId:  m.installPvId,
		installId:    m.installId,
		spaceId:      m.spaceId,
		capabilityId: m.capabilityId,
		db:           m.db,
	}, nil
}

func (m *MigratorCapability) Close() error {
	return nil
}

func (m *MigratorCapability) performMigration(folder string) error {

	qq.Println(
		"@performMigration/1",
		"Starting migration process for installId:", m.installId,
		"folder:", folder,
		"installPvId:", m.installPvId,
		"capabilityId:", m.capabilityId,
	)

	migFolder := m.folder

	if migFolder == "" {
		migFolder = folder
	}

	pkgFileOps := m.builder.app.Database().GetPackageFileOps()

	files, err := pkgFileOps.ListFiles(m.installPvId, migFolder)
	if err != nil {
		return err
	}

	qq.Println("@performMigration/3", "files found:", len(files))

	if len(files) == 0 {
		return nil
	}

	// Filter only .sql files
	sqlFiles := []dbmodels.FileMeta{}
	for _, file := range files {
		if !file.IsFolder && strings.HasSuffix(strings.ToLower(file.Name), ".sql") {
			sqlFiles = append(sqlFiles, file)
		}
	}

	qq.Println("@performMigration/4", "SQL files found:", len(sqlFiles))

	if len(sqlFiles) == 0 {
		return nil
	}

	// Sort files by name, alphabetically
	sort.Slice(sqlFiles, func(i, j int) bool {
		return sqlFiles[i].Name < sqlFiles[j].Name
	})

	qq.Println("@performMigration/5", "SQL files sorted")

	// Get list of already executed migrations
	executedMigrations, err := m.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	qq.Println("@performMigration/6", "Executed migrations retrieved:", len(executedMigrations))

	// Execute each migration that hasn't been run yet
	for _, file := range sqlFiles {

		qq.Println("@performMigration/7", "Processing file:", file.Name)

		migrationKey := getMigrationKey(migFolder, file)

		qq.Println("@performMigration/8", "Migration key:", migrationKey)

		// Skip if already executed
		if executedMigrations[migrationKey] {

			qq.Println("@performMigration/9", "Already executed, skipping:", file.Name)

			continue
		}

		qq.Println("@performMigration/10", "Executing migration:", file.Name)

		content, err := pkgFileOps.GetFileContentByPath(m.installPvId, file.Path, file.Name)
		if err != nil {

			qq.Println("@performMigration/11", "Failed to read file content:", file.Name, "error:", err)

			return fmt.Errorf("failed to read migration file %s: %w", file.Name, err)
		}

		qq.Println("@performMigration/12", "File content read successfully:", file.Name)

		// Execute SQL
		sqlContent := string(content)
		if strings.TrimSpace(sqlContent) == "" {

			qq.Println("@performMigration/13", "Empty SQL content, skipping execution but marking as executed:", file.Name)

			// Skip empty files but mark as executed
			if err := m.markMigrationExecuted(migrationKey, file.Name); err != nil {
				return fmt.Errorf("failed to mark migration as executed: %w", err)
			}

			qq.Println("@performMigration/14", "Migration marked as executed:", file.Name)
			continue
		}

		qq.Println("@performMigration/15", "Executing SQL content:", file.Name)

		_, err = m.db.Exec(sqlContent)
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file.Name, err)
		}

		qq.Println("@performMigration/16", "SQL content executed successfully:", file.Name)

		// Mark migration as executed
		if err := m.markMigrationExecuted(migrationKey, file.Name); err != nil {
			return fmt.Errorf("failed to mark migration as executed: %w", err)
		}
	}

	return nil
}

func (m *MigratorCapability) getExecutedMigrations() (map[string]bool, error) {
	executed := make(map[string]bool)

	spaceKVOps := m.builder.app.Database().GetSpaceKVOps()

	migs, err := spaceKVOps.QuerySpaceKV(m.installId, map[any]any{"group": "migrations"}, 0, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get executed migrations: %w", err)
	}

	for _, mig := range migs {
		executed[mig.Key] = true
	}

	return executed, nil
}

func (m *MigratorCapability) markMigrationExecuted(migrationKey, fileName string) error {
	spaceKVOps := m.builder.app.Database().GetSpaceKVOps()
	err := spaceKVOps.AddSpaceKV(m.installId, &dbmodels.SpaceKV{
		Group: "migrations",
		Key:   migrationKey,
		Value: fileName,
	})
	if err != nil {
		return fmt.Errorf("failed to mark migration as executed: %w", err)
	}

	return nil
}

func getMigrationKey(migFolder string, file dbmodels.FileMeta) string {
	// Use a combination of path and filename as the unique key
	path := file.Path
	if path == "" {
		path = migFolder
	}
	return fmt.Sprintf("%s/%s", path, file.Name)
}
