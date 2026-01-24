package cdc

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"text/template"
)

const TemplateTable = `
CREATE TABLE IF NOT EXISTS {{.TableName}} (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
	record_id INTEGER NOT NULL,
	operation INTEGER NOT NULL, -- 0: insert, 1: update, 2: delete
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

func EnsureCDC(db *sql.DB) error {
	// list all tables in the database
	tableNames, err := getTableNames(db)
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}

	var tables []string
	existingCDCTables := make(map[string]bool)

	for _, tableName := range tableNames {
		// skip if it has postfix "__cdc"
		if strings.HasSuffix(tableName, "__cdc") {
			continue
		}

		if tableName == "CDCMeta" {
			continue
		}

		// check if there is {table_name}_cdc table
		cdcTableName := tableName + "_cdc"
		var exists int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", cdcTableName).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check CDC table existence: %w", err)
		}

		if exists > 0 {
			existingCDCTables[tableName] = true
			continue
		}

		tables = append(tables, tableName)
	}

	// Parse the template
	tmpl, err := template.New("cdc_table").Parse(TemplateTable)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// for each table, create a table with the name {table_name}_cdc
	for _, tableName := range tables {
		cdcTableName := tableName + "_cdc"

		// Detect primary key column
		pkColumn, err := getPrimaryKeyColumn(db, tableName)
		if err != nil {
			return fmt.Errorf("failed to get primary key for table %s: %w", tableName, err)
		}

		// Create CDC table
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, map[string]string{"TableName": cdcTableName}); err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", tableName, err)
		}

		if _, err := db.Exec(buf.String()); err != nil {
			return fmt.Errorf("failed to create CDC table %s: %w", cdcTableName, err)
		}

		// create a trigger on the table for each operation
		// Trigger for INSERT
		insertTrigger := fmt.Sprintf(`
			CREATE TRIGGER IF NOT EXISTS %s_insert_cdc
			AFTER INSERT ON %s
			BEGIN
				INSERT INTO %s (record_id, operation) VALUES (NEW.%s, 0);
			END;
		`, tableName, tableName, cdcTableName, pkColumn)

		if _, err := db.Exec(insertTrigger); err != nil {
			return fmt.Errorf("failed to create INSERT trigger for %s: %w", tableName, err)
		}

		// Trigger for UPDATE
		updateTrigger := fmt.Sprintf(`
			CREATE TRIGGER IF NOT EXISTS %s_update_cdc
			AFTER UPDATE ON %s
			BEGIN
				INSERT INTO %s (record_id, operation) VALUES (NEW.%s, 1);
			END;
		`, tableName, tableName, cdcTableName, pkColumn)

		if _, err := db.Exec(updateTrigger); err != nil {
			return fmt.Errorf("failed to create UPDATE trigger for %s: %w", tableName, err)
		}

		// Trigger for DELETE
		deleteTrigger := fmt.Sprintf(`
			CREATE TRIGGER IF NOT EXISTS %s_delete_cdc
			AFTER DELETE ON %s
			BEGIN
				INSERT INTO %s (record_id, operation) VALUES (OLD.%s, 2);
			END;
		`, tableName, tableName, cdcTableName, pkColumn)

		if _, err := db.Exec(deleteTrigger); err != nil {
			return fmt.Errorf("failed to create DELETE trigger for %s: %w", tableName, err)
		}

		// Insert or update a record in CDCMeta table
		// If table already has records, set cdc_start_id to max rowid
		if err := ensureCDCMeta(db, tableName, pkColumn); err != nil {
			return fmt.Errorf("failed to ensure CDCMeta for table %s: %w", tableName, err)
		}
	}

	return nil
}

// ensureCDCMeta ensures a CDCMeta record exists for the given table.
// If the table already has records, it sets cdc_start_id to the max primary key value.
func ensureCDCMeta(db *sql.DB, tableName string, pkColumn string) error {
	// Check if CDCMeta record already exists
	var existingID int
	err := db.QueryRow("SELECT id FROM CDCMeta WHERE table_name = ?", tableName).Scan(&existingID)
	recordExists := err == nil

	// Get max primary key value from the table
	var maxID sql.NullInt64
	quotedTableName := fmt.Sprintf(`"%s"`, strings.ReplaceAll(tableName, `"`, `""`))
	quotedPkColumn := fmt.Sprintf(`"%s"`, strings.ReplaceAll(pkColumn, `"`, `""`))
	err = db.QueryRow(fmt.Sprintf("SELECT MAX(%s) FROM %s", quotedPkColumn, quotedTableName)).Scan(&maxID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get max id for table %s: %w", tableName, err)
	}

	cdcStartID := int64(0)
	if maxID.Valid && maxID.Int64 > 0 {
		// Table has existing records, set cdc_start_id to max rowid
		cdcStartID = maxID.Int64
	}

	if recordExists {
		// Update existing record (only if cdc_start_id is 0, meaning it hasn't been set yet)
		_, err = db.Exec(`
			UPDATE CDCMeta 
			SET cdc_start_id = ? 
			WHERE table_name = ? AND cdc_start_id = 0
		`, cdcStartID, tableName)
		if err != nil {
			return fmt.Errorf("failed to update CDCMeta for table %s: %w", tableName, err)
		}
	} else {
		// Insert new record
		_, err = db.Exec(`
			INSERT INTO CDCMeta (table_name, cdc_start_id, current_cdc_id, gc_max_records, last_gc_at, extrameta)
			VALUES (?, ?, 0, 0, 0, '{}')
		`, tableName, cdcStartID)
		if err != nil {
			return fmt.Errorf("failed to insert CDCMeta for table %s: %w", tableName, err)
		}
	}

	return nil
}

func DropCDC(db *sql.DB) error {
	tableNames, err := getTableNames(db)
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}

	for _, tableName := range tableNames {
		if strings.HasSuffix(tableName, "_cdc") {
			continue
		}

		// drop triggers on the table
		if _, err := db.Exec("DROP TRIGGER IF EXISTS ?", tableName+"_insert_cdc"); err != nil {
			return fmt.Errorf("failed to drop INSERT trigger for %s: %w", tableName, err)
		}
		if _, err := db.Exec("DROP TRIGGER IF EXISTS ?", tableName+"_update_cdc"); err != nil {
			return fmt.Errorf("failed to drop UPDATE trigger for %s: %w", tableName, err)
		}
		if _, err := db.Exec("DROP TRIGGER IF EXISTS ?", tableName+"_delete_cdc"); err != nil {
			return fmt.Errorf("failed to drop DELETE trigger for %s: %w", tableName, err)
		}

		cdcTableName := tableName + "_cdc"
		if _, err := db.Exec("DROP TABLE IF EXISTS ?", cdcTableName); err != nil {
			return fmt.Errorf("failed to drop CDC table %s: %w", cdcTableName, err)
		}

	}

	// truncate CDCMeta table
	if _, err := db.Exec("TRUNCATE TABLE CDCMeta"); err != nil {
		return fmt.Errorf("failed to truncate CDCMeta table: %w", err)
	}

	return nil

}
