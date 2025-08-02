package blua

import (
	"database/sql"

	"github.com/blue-monads/turnix/backend/services/database"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/rs/zerolog"
)

type BuilderOption struct {
	App xtypes.App
	DB  *database.DB
	SQL SQL

	Logger zerolog.Logger
}

type SQL interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) (*sql.Row, error)
	RunDDL(query string) error
	RunMigration(pid int64, name string, query string) error
	ListMigration(pid int64) ([]string, error)
	RemoveMigration(pid int64, name string) error

	// Transaction
	Begin() (*sql.Tx, error)
}

type Builder func(opt BuilderOption) (*XSpaceType, error)

type XSpaceField struct {
	Name       string   `json:"name"`
	KeyName    string   `json:"key_name"`
	Ftype      string   `json:"ftype"`
	Disabled   bool     `json:"disabled"`
	Options    []string `json:"options"`
	IsRequired bool     `json:"is_required"`
}

type XSpaceType struct {
	Name    string
	Slug    string
	Info    string
	Version string

	Fields         []XSpaceField
	EventHandlers  map[string]any
	AssetsProvider any
}
