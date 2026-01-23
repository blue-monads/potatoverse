package staticseeder

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

type StaticSeederCapability struct {
	seedFolder   string
	builder      *StaticSeederBuilder
	installId    int64
	spaceId      int64
	capabilityId int64
	db           datahub.DBLowOps
}

func (s *StaticSeederCapability) performSeed() error {
	pkgFileOps := s.builder.app.Database().GetPackageFileOps()

	// List all files in the seed folder
	files, err := pkgFileOps.ListFiles(s.installId, s.seedFolder)
	if err != nil {
		return fmt.Errorf("failed to list files in seed folder %s: %w", s.seedFolder, err)
	}

	if len(files) == 0 {
		return nil
	}

	// Filter only JSON files (not folders)
	jsonFiles := []dbmodels.FileMeta{}
	for _, file := range files {
		if !file.IsFolder && strings.HasSuffix(strings.ToLower(file.Name), ".json") {
			jsonFiles = append(jsonFiles, file)
		}
	}

	if len(jsonFiles) == 0 {
		return nil
	}

	// Sort files by name, alphabetically
	sort.Slice(jsonFiles, func(i, j int) bool {
		return jsonFiles[i].Name < jsonFiles[j].Name
	})

	// Process each file in alphabetical order
	for _, file := range jsonFiles {
		if err := s.processSeedFile(file.Path, file.Name); err != nil {
			return fmt.Errorf("failed to process seed file %s/%s: %w", file.Path, file.Name, err)
		}
	}

	return nil
}

func (s *StaticSeederCapability) processSeedFile(dir, filename string) error {
	pkgFileOps := s.builder.app.Database().GetPackageFileOps()

	content, err := pkgFileOps.GetFileContentByPath(s.installId, dir, filename)
	if err != nil {
		return fmt.Errorf("failed to read seed file %s/%s: %w", dir, filename, err)
	}

	// Parse JSON into StaticSeederStruct array
	var seedStructs []StaticSeederStruct
	if err := json.Unmarshal(content, &seedStructs); err != nil {
		return fmt.Errorf("failed to parse seed file %s/%s: %w", dir, filename, err)
	}

	// Insert data for each table
	for _, seedStruct := range seedStructs {
		if err := s.insertTableData(seedStruct); err != nil {
			return fmt.Errorf("failed to insert data for table %s: %w", seedStruct.TableName, err)
		}
	}

	return nil
}

func (s *StaticSeederCapability) insertTableData(seedStruct StaticSeederStruct) error {
	// Insert each row of data
	for _, rowData := range seedStruct.Data {
		_, err := s.db.Insert(seedStruct.TableName, rowData)
		if err != nil {
			return fmt.Errorf("failed to insert row into %s: %w", seedStruct.TableName, err)
		}
	}

	return nil
}

func (s *StaticSeederCapability) Handle(ctx *gin.Context) {}

func (s *StaticSeederCapability) ListActions() ([]string, error) {
	return []string{"seed"}, nil
}

func (s *StaticSeederCapability) Execute(name string, params lazydata.LazyData) (any, error) {
	switch name {
	case "seed":
		if err := s.performSeed(); err != nil {
			return nil, fmt.Errorf("seeding failed: %w", err)
		}
		return map[string]any{"status": "success", "message": "seeding completed"}, nil

	default:
		return nil, fmt.Errorf("invalid action: %s", name)
	}
}

func (s *StaticSeederCapability) Reload(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {
	return &StaticSeederCapability{
		seedFolder:   s.seedFolder,
		builder:      s.builder,
		installId:    s.installId,
		spaceId:      s.spaceId,
		capabilityId: s.capabilityId,
		db:           s.db,
	}, nil
}

func (s *StaticSeederCapability) Close() error {
	return nil
}
