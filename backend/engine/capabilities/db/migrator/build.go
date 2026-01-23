package migrator

import (
	"fmt"

	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

var (
	Name         = "migrator"
	Icon         = ""
	OptionFields = []xcapability.CapabilityOptionField{
		{
			Name:        "Migration Folder",
			Key:         "folder",
			Description: "Folder path containing migration SQL files (e.g., 'migrations' or 'db/migrations')",
			Type:        "text",
			Default:     "migrations",
			Required:    true,
		},
	}
)

func init() {
	registry.RegisterCapability(Name, xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			appTyped := app.(xtypes.App)
			return &MigratorBuilder{app: appTyped}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	})
}

type MigratorBuilder struct {
	app xtypes.App
}

type MigratorOptions struct {
	Folder string `json:"folder"`
}

func (b *MigratorBuilder) Name() string {
	return Name
}

func (b *MigratorBuilder) Build(handle xcapability.XCapabilityHandle) (xcapability.Capability, error) {
	model := handle.GetModel()

	var opts MigratorOptions
	if err := handle.GetOptions(&opts); err != nil {
		return nil, fmt.Errorf("failed to parse options: %w", err)
	}

	// Default folder if not specified
	folder := opts.Folder
	if folder == "" {
		folder = "migrations"
	}

	db := b.app.Database().GetLowPackageDBOps(model.InstallID)

	capability := &MigratorCapability{
		folder:       folder,
		builder:      b,
		installId:    model.InstallID,
		spaceId:      model.SpaceID,
		capabilityId: model.ID,
		db:           db,
	}

	// Run migrations automatically on build
	if err := capability.performMigration(); err != nil {
		return nil, fmt.Errorf("failed to run initial migrations: %w", err)
	}

	return capability, nil
}

func (b *MigratorBuilder) Serve(ctx *gin.Context) {}

func (b *MigratorBuilder) GetDebugData() map[string]any {
	return map[string]any{
		"name": Name,
	}
}
