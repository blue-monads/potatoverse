package autocdc

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
)

// AutoCapture captures changes in sqlite database tables.

type TableSchema struct {
	Type     string  `json:"type" db:"type,omitempty"`
	Name     string  `json:"name" db:"name"`
	TblName  string  `json:"tbl_name" db:"tbl_name"`
	RootPage int64   `json:"rootpage" db:"rootpage"`
	Sql      *string `json:"sql" db:"sql"`
}

type CaptureOptions struct {
	Logger *slog.Logger
}

type AutoCapture struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewAutoCapture(opts CaptureOptions) *AutoCapture {
	return &AutoCapture{
		db:     nil,
		logger: opts.Logger,
	}
}

func (ac *AutoCapture) SetDB(db *sql.DB) {
	ac.db = db
}

func (ac *AutoCapture) Init() error {

	tables, err := ac.getTableSchemas()
	if err != nil {
		ac.logger.Error("Failed to get table schemas", "error", err)
		return err
	}

	for _, table := range tables {
		ac.logger.Info("Found table schema", "table", table.Name)

		if strings.HasPrefix(table.TblName, "sqlite_") {
			ac.logger.Debug("Skipping sqlite internal table", "table", table.TblName)
			continue
		}

		if strings.HasPrefix(table.TblName, "zcdc_") {
			ac.logger.Debug("Skipping existing internal cdc table", "table", table.TblName)
			continue
		}

		if strings.HasPrefix(table.TblName, "zbuddy_") {
			ac.logger.Debug("Skipping buddy table", "table", table.TblName)
			continue
		}

		columns, err := ac.getTableColumns(table.TblName)
		if err != nil {
			ac.logger.Error("Failed to get table columns", "table", table.TblName, "error", err)
			return err
		}

		ddls := ac.buildCaptureDDL2(table.TblName, columns)

		// var sbuf strings.Builder

		for _, tmpl := range ddls {
			_, err := ac.db.Exec(tmpl)
			if err != nil {
				ac.logger.Error("Failed to execute DDL", "ddl", tmpl, "error", err)
				return err
			}
			// sbuf.WriteString(tmpl + "\n")
			ac.logger.Debug("Executed DDL", "ddl", tmpl)
		}

		//os.WriteFile("triggers.sql", []byte(sbuf.String()), 0644)

	}

	return nil
}

func (ac *AutoCapture) TriggerRebuild(tables []string) error {
	ac.logger.Info("Triggering rebuild of some tables", "tables", tables)

	for _, table := range tables {
		ac.logger.Info("Rebuilding CDC triggers for table", "table", table)
		columns, err := ac.getTableColumns(table)
		if err != nil {
			ac.logger.Error("Failed to get table columns", "table", table, "error", err)
			return err
		}
		ddls := ac.buildCaptureDDL2(table, columns)
		for _, tmpl := range ddls {
			_, err := ac.db.Exec(tmpl)
			if err != nil {
				ac.logger.Error("Failed to execute DDL", "ddl", tmpl, "error", err)
				return err
			}
			ac.logger.Debug("Executed DDL", "ddl", tmpl)
		}
	}

	return nil
}

// private

func (ac *AutoCapture) getTableSchemas() ([]TableSchema, error) {
	rows, err := ac.db.Query("SELECT type, name, tbl_name, rootpage, sql FROM sqlite_master WHERE type IN ('table')")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schemas []TableSchema
	for rows.Next() {
		var schema TableSchema
		if err := rows.Scan(&schema.Type, &schema.Name, &schema.TblName, &schema.RootPage, &schema.Sql); err != nil {
			return nil, err
		}
		schemas = append(schemas, schema)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schemas, nil
}

func (ac *AutoCapture) getTableColumns(tableName string) ([]string, error) {
	rows, err := ac.db.Query("PRAGMA table_info(" + tableName + ")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt_value sql.NullString

		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			return nil, err
		}
		columns = append(columns, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

func queryRowData(db *sql.DB, tableName string, id int64) (map[string]any, error) {
	selectQuery := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", tableName)
	rows, err := db.Query(selectQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query row: %v", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %v", err)
	}

	if !rows.Next() {
		return nil, fmt.Errorf("no row found with id %d", id)
	}

	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	err = rows.Scan(valuePtrs...)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %v", err)
	}

	result := make(map[string]any)
	for i, col := range columns {
		result[col] = values[i]
	}

	return result, nil
}
