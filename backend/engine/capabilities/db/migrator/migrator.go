package migrator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

type MigratorCapability struct {
	folder       string
	builder      *MigratorBuilder
	installId    int64
	spaceId      int64
	capabilityId int64
	db           datahub.DBLowOps
}

func (m *MigratorCapability) performMigration() error {

	pkgFileOps := m.builder.app.Database().GetPackageFileOps()

	files, err := pkgFileOps.ListFiles(m.installId, m.folder)
	if err != nil {
		return err
	}

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

	if len(sqlFiles) == 0 {
		return nil
	}

	// Sort files by name, alphabetically
	sort.Slice(sqlFiles, func(i, j int) bool {
		return sqlFiles[i].Name < sqlFiles[j].Name
	})

	// Get list of already executed migrations
	executedMigrations, err := m.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	// Execute each migration that hasn't been run yet
	for _, file := range sqlFiles {
		migrationKey := m.getMigrationKey(file)

		// Skip if already executed
		if executedMigrations[migrationKey] {

			continue
		}

		content, err := pkgFileOps.GetFileContentByPath(m.installId, file.Path, file.Name)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file.Name, err)
		}

		// Execute SQL
		sqlContent := string(content)
		if strings.TrimSpace(sqlContent) == "" {
			// Skip empty files but mark as executed
			if err := m.markMigrationExecuted(migrationKey, file.Name); err != nil {
				return fmt.Errorf("failed to mark migration as executed: %w", err)
			}
			continue
		}

		_, err = m.db.Exec(sqlContent)
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file.Name, err)
		}

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

func (m *MigratorCapability) getMigrationKey(file dbmodels.FileMeta) string {
	// Use a combination of path and filename as the unique key
	path := file.Path
	if path == "" {
		path = m.folder
	}
	return fmt.Sprintf("%s/%s", path, file.Name)
}

func (m *MigratorCapability) Handle(ctx *gin.Context) {}

func (m *MigratorCapability) ListActions() ([]string, error) {
	return []string{"run_migrations", "list_migrations"}, nil
}

func (m *MigratorCapability) Execute(name string, params lazydata.LazyData) (any, error) {
	switch name {
	case "run_migrations":
		if err := m.performMigration(); err != nil {
			return nil, fmt.Errorf("migration failed: %w", err)
		}
		return map[string]any{"status": "success", "message": "migrations completed"}, nil

	case "list_migrations":
		executed, err := m.getExecutedMigrations()
		if err != nil {
			return nil, err
		}

		pkgFileOps := m.builder.app.Database().GetPackageFileOps()
		files, err := pkgFileOps.ListFiles(m.installId, m.folder)
		if err != nil {
			return nil, err
		}

		migrations := []map[string]any{}
		for _, file := range files {
			if !file.IsFolder && strings.HasSuffix(strings.ToLower(file.Name), ".sql") {
				migrationKey := m.getMigrationKey(file)
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
		installId:    m.installId,
		spaceId:      m.spaceId,
		capabilityId: m.capabilityId,
		db:           m.db,
	}, nil
}

func (m *MigratorCapability) Close() error {
	return nil
}
