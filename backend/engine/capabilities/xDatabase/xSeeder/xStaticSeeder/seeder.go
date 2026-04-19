package staticseeder

import (
	"encoding/json"
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

type StaticSeederCapability struct {
	seedFolder   string
	builder      *StaticSeederBuilder
	installId    int64
	installPvId  int64
	spaceId      int64
	capabilityId int64
	db           datahub.DBLowOps
}

func (s *StaticSeederCapability) performSeed(folder string) error {
	pkgFileOps := s.builder.app.Database().GetPackageFileOps()

	seedFolder := s.seedFolder

	if seedFolder != "" {
		seedFolder = folder
	}

	qq.Println("@performSeed/1", "Starting seed process for installPvId:", s.installPvId, "folder:", seedFolder, "installId:", s.installId)

	// List all files in the seed folder
	files, err := pkgFileOps.ListFiles(s.installPvId, seedFolder)
	if err != nil {
		qq.Println("@performSeed/2", "Failed to list files:", err)
		return fmt.Errorf("failed to list files in seed folder %s: %w", seedFolder, err)
	}

	qq.Println("@performSeed/3", "files found:", len(files))

	if len(files) == 0 {
		qq.Println("@performSeed/4", "No files found in seed folder")
		return nil
	}

	// Filter only JSON files (not folders)
	jsonFiles := []dbmodels.FileMeta{}
	for _, file := range files {
		if !file.IsFolder && strings.HasSuffix(strings.ToLower(file.Name), ".json") {
			jsonFiles = append(jsonFiles, file)
		}
	}

	qq.Println("@performSeed/5", "JSON files found:", len(jsonFiles))

	if len(jsonFiles) == 0 {
		qq.Println("@performSeed/6", "No JSON files found")
		return nil
	}

	// Sort files by name, alphabetically
	sort.Slice(jsonFiles, func(i, j int) bool {
		return jsonFiles[i].Name < jsonFiles[j].Name
	})

	qq.Println("@performSeed/7", "JSON files sorted")

	// Process each file in alphabetical order
	for _, file := range jsonFiles {
		qq.Println("@performSeed/8", "Processing file:", file.Name)
		if err := s.processSeedFile(file.Path, file.Name); err != nil {
			qq.Println("@performSeed/9", "Failed to process file:", file.Name, "error:", err)
			return fmt.Errorf("failed to process seed file %s/%s: %w", file.Path, file.Name, err)
		}
		qq.Println("@performSeed/10", "Successfully processed file:", file.Name)
	}

	qq.Println("@performSeed/11", "Seeding completed successfully")

	return nil
}

func (s *StaticSeederCapability) processSeedFile(dir, filename string) error {
	qq.Println("@processSeedFile/1", "Processing seed file:", dir, filename)

	pkgFileOps := s.builder.app.Database().GetPackageFileOps()

	content, err := pkgFileOps.GetFileContentByPath(s.installPvId, dir, filename)
	if err != nil {
		qq.Println("@processSeedFile/2", "Failed to read file content:", err)
		return fmt.Errorf("failed to read seed file %s/%s: %w", dir, filename, err)
	}

	qq.Println("@processSeedFile/3", "File content read, length:", len(content))

	// Parse JSON into StaticSeederStruct array
	var seedStructs []StaticSeederStruct
	if err := json.Unmarshal(content, &seedStructs); err != nil {
		qq.Println("@processSeedFile/4", "Failed to parse JSON:", err)
		return fmt.Errorf("failed to parse seed file %s/%s: %w", dir, filename, err)
	}

	qq.Println("@processSeedFile/5", "Parsed seed structs:", len(seedStructs))

	// Insert data for each table
	for _, seedStruct := range seedStructs {
		qq.Println("@processSeedFile/6", "Inserting data for table:", seedStruct.TableName)
		if err := s.insertTableData(seedStruct); err != nil {
			qq.Println("@processSeedFile/7", "Failed to insert data for table:", seedStruct.TableName, "error:", err)
			return fmt.Errorf("failed to insert data for table %s: %w", seedStruct.TableName, err)
		}
		qq.Println("@processSeedFile/8", "Successfully inserted data for table:", seedStruct.TableName)
	}

	qq.Println("@processSeedFile/9", "Processed seed file successfully")

	return nil
}

func (s *StaticSeederCapability) insertTableData(seedStruct StaticSeederStruct) error {
	qq.Println("@insertTableData/1", "Inserting", len(seedStruct.Data), "rows into table:", seedStruct.TableName)

	// Insert each row of data
	for i, rowData := range seedStruct.Data {
		qq.Println("@insertTableData/2", "Inserting row", i+1, "data:", rowData)
		_, err := s.db.Insert(seedStruct.TableName, rowData)
		if err != nil {
			qq.Println("@insertTableData/3", "Failed to insert row:", err)
			return fmt.Errorf("failed to insert row into %s: %w", seedStruct.TableName, err)
		}
		qq.Println("@insertTableData/4", "Successfully inserted row", i+1)
	}

	qq.Println("@insertTableData/5", "Inserted all rows for table:", seedStruct.TableName)

	return nil
}

func (s *StaticSeederCapability) Handle(ctx *gin.Context) {}

func (s *StaticSeederCapability) ListActions() ([]string, error) {
	return []string{"seed"}, nil
}

func (s *StaticSeederCapability) Execute(name string, params lazydata.LazyData) (any, error) {
	switch name {
	case "seed":
		folder := params.GetFieldAsString("seed_folder")

		if err := s.performSeed(folder); err != nil {
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
		installPvId:  s.installPvId,
		spaceId:      s.spaceId,
		capabilityId: s.capabilityId,
		db:           s.db,
	}, nil
}

func (s *StaticSeederCapability) Close() error {
	return nil
}
