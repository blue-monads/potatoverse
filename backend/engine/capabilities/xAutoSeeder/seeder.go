package autoseeder

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/gin-gonic/gin"
)

type AutoseederCapability struct {
	autoseederStructFile string
	builder              *AutoseederBuilder
	installId            int64
	spaceId              int64
	capabilityId         int64
	db                   datahub.DBLowOps
}

func (s *AutoseederCapability) performSeed() error {
	// Note: Global rand functions are used for null probability checks
	// In Go 1.20+, they are auto-seeded, but we seed explicitly for older versions
	rand.Seed(time.Now().UnixNano())

	pkgFileOps := s.builder.app.Database().GetPackageFileOps()

	dir, filename := filepath.Split(s.autoseederStructFile)

	content, err := pkgFileOps.GetFileContentByPath(s.installId, dir, filename)
	if err != nil {
		return fmt.Errorf("failed to read seed file %s: %w", s.autoseederStructFile, err)
	}

	// Parse JSON into SeedStruct array
	var seedStructs []SeedStruct
	if err := json.Unmarshal(content, &seedStructs); err != nil {
		return fmt.Errorf("failed to parse seed file: %w", err)
	}

	// Seed each table
	for _, seedStruct := range seedStructs {
		if err := s.seedTable(seedStruct); err != nil {
			return fmt.Errorf("failed to seed table %s: %w", seedStruct.TableName, err)
		}
	}

	return nil
}

func (s *AutoseederCapability) seedTable(seedStruct SeedStruct) error {
	// Generate and insert rows
	for i := 0; i < seedStruct.Rows; i++ {
		rowData := make(map[string]any)

		for _, col := range seedStruct.Columns {
			value := s.generateValue(col)
			rowData[col.ColumnName] = value
		}

		_, err := s.db.Insert(seedStruct.TableName, rowData)
		if err != nil {
			return fmt.Errorf("failed to insert row into %s: %w", seedStruct.TableName, err)
		}
	}

	return nil
}

func (s *AutoseederCapability) generateValue(col SeedColumn) any {
	// Check if we should generate null
	if !col.NotNull && col.NullProbability > 0 {
		if rand.Float64() < col.NullProbability {
			return nil
		}
	}

	// Handle default values (for enums)
	if len(col.DefaultValues) > 0 {
		return col.DefaultValues[rand.Intn(len(col.DefaultValues))]
	}

	// Generate value based on data type
	switch col.DataType {
	case "string", "varchar", "text":
		return gofakeit.Word()
	case "int", "integer", "int64":
		if col.RangeMin != 0 || col.RangeMax != 0 {
			return gofakeit.IntRange(col.RangeMin, col.RangeMax)
		}
		return gofakeit.Int64()
	case "float", "float64", "decimal", "numeric":
		if col.RangeMin != 0 || col.RangeMax != 0 {
			return gofakeit.Float64Range(float64(col.RangeMin), float64(col.RangeMax))
		}
		return gofakeit.Float64()
	case "bool", "boolean":
		return gofakeit.Bool()
	case "date", "timestamp", "datetime":
		if col.RangeMin != 0 || col.RangeMax != 0 {
			minTime := time.Unix(int64(col.RangeMin), 0)
			maxTime := time.Unix(int64(col.RangeMax), 0)
			return gofakeit.DateRange(minTime, maxTime)
		}
		return gofakeit.Date()
	case "email":
		return gofakeit.Email()
	case "url":
		return gofakeit.URL()
	case "phone":
		return gofakeit.Phone()
	case "uuid":
		return gofakeit.UUID()
	case "ipv4":
		return gofakeit.IPv4Address()
	case "ipv6":
		return gofakeit.IPv6Address()
	case "name":
		return gofakeit.Name()
	case "firstname":
		return gofakeit.FirstName()
	case "lastname":
		return gofakeit.LastName()
	case "username":
		return gofakeit.Username()
	case "password":
		return gofakeit.Password(true, true, true, false, false, 10)
	case "address":
		return gofakeit.Address().Address
	case "city":
		return gofakeit.City()
	case "country":
		return gofakeit.Country()
	case "zip":
		return gofakeit.Zip()
	case "company":
		return gofakeit.Company()
	case "jobtitle":
		return gofakeit.JobTitle()
	case "sentence":
		return gofakeit.Sentence(10)
	case "paragraph":
		return gofakeit.Paragraph(3, 5, 10, " ")
	case "product_name":
		return gofakeit.ProductName()
	case "product_description":
		return gofakeit.ProductDescription()
	case "product_category":
		return gofakeit.ProductCategory()
	case "color":
		return gofakeit.Color()
	case "hex_color":
		return gofakeit.HexColor()
	case "rgb_color":
		return gofakeit.RGBColor()

	default:

		// For unknown types, return a generic string
		// Users can specify custom gofakeit function names in DataType
		// For now, return a placeholder string
		return fmt.Sprintf("unknown_type_%s", col.DataType)
	}
}

func (s *AutoseederCapability) Handle(ctx *gin.Context) {}

func (s *AutoseederCapability) ListActions() ([]string, error) {
	return []string{"seed"}, nil
}

func (s *AutoseederCapability) Execute(name string, params lazydata.LazyData) (any, error) {
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

func (s *AutoseederCapability) Reload(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {
	return &AutoseederCapability{
		autoseederStructFile: s.autoseederStructFile,
		builder:              s.builder,
		installId:            s.installId,
		spaceId:              s.spaceId,
		capabilityId:         s.capabilityId,
		db:                   s.db,
	}, nil
}

func (s *AutoseederCapability) Close() error {
	return nil
}
