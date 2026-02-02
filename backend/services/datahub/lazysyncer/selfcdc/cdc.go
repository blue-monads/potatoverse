package selfcdc

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"text/template"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazymodel"
)

func EnsureCDC(db *sql.DB) (int, error) {
	// list all tables in the database
	tableNames, err := getTableNames(db)
	if err != nil {
		return 0, fmt.Errorf("failed to list tables: %w", err)
	}

	var tables []string
	existingCDCTables := make(map[string]bool)

	for _, tableName := range tableNames {
		// skip if it has postfix "__cdc"
		if strings.HasSuffix(tableName, "__cdc") {
			continue
		}

		if slices.Contains(SkipTables, tableName) {
			continue
		}

		// check if there is {table_name}__cdc table
		cdcTableName := tableName + "__cdc"
		var exists int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", cdcTableName).Scan(&exists)
		if err != nil {
			return 0, fmt.Errorf("failed to check CDC table existence: %w", err)
		}

		if exists > 0 {
			existingCDCTables[tableName] = true
			continue
		}

		tables = append(tables, tableName)
	}

	// Parse the template
	tmpl, err := template.New("cdc_table").Parse(lazymodel.CDCTableTemplate)
	if err != nil {
		return 0, fmt.Errorf("failed to parse template: %w", err)
	}

	newEntries := 0

	// for each table, create a table with the name {table_name}__cdc
	for _, tableName := range tables {
		cdcTableName := tableName + "__cdc"

		// Detect primary key column
		pkColumn, err := getPrimaryKeyColumn(db, tableName)
		if err != nil {
			return 0, fmt.Errorf("failed to get primary key for table %s: %w", tableName, err)
		}

		// Create CDC table
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, map[string]string{"TableName": cdcTableName}); err != nil {
			return 0, fmt.Errorf("failed to execute template for %s: %w", tableName, err)
		}

		if _, err := db.Exec(buf.String()); err != nil {
			return 0, fmt.Errorf("failed to create CDC table %s: %w", cdcTableName, err)
		}

		// create a trigger on the table for each operation
		// Trigger for INSERT
		insertTrigger := fmt.Sprintf(`
			CREATE TRIGGER IF NOT EXISTS %s_insert__cdc
			AFTER INSERT ON %s
			BEGIN
				INSERT INTO %s (record_id, operation) VALUES (NEW.%s, 0);
			END;
		`, tableName, tableName, cdcTableName, pkColumn)

		if _, err := db.Exec(insertTrigger); err != nil {
			return 0, fmt.Errorf("failed to create INSERT trigger for %s: %w", tableName, err)
		}

		// Trigger for UPDATE
		updateTrigger := fmt.Sprintf(`
			CREATE TRIGGER IF NOT EXISTS %s_update__cdc
			AFTER UPDATE ON %s
			BEGIN
				INSERT INTO %s (record_id, operation) VALUES (NEW.%s, 1);
			END;
		`, tableName, tableName, cdcTableName, pkColumn)

		if _, err := db.Exec(updateTrigger); err != nil {
			return 0, fmt.Errorf("failed to create UPDATE trigger for %s: %w", tableName, err)
		}

		// Trigger for DELETE
		deleteTrigger := fmt.Sprintf(`
			CREATE TRIGGER IF NOT EXISTS %s_delete__cdc
			AFTER DELETE ON %s
			BEGIN
				INSERT INTO %s (record_id, operation) VALUES (OLD.%s, 2);
			END;
		`, tableName, tableName, cdcTableName, pkColumn)

		if _, err := db.Exec(deleteTrigger); err != nil {
			return 0, fmt.Errorf("failed to create DELETE trigger for %s: %w", tableName, err)
		}

		// Insert or update a record in CDCMeta table
		// If table already has records, set cdc_start_id to max rowid
		if err := ensureCDCMeta(db, tableName, pkColumn); err != nil {
			return 0, fmt.Errorf("failed to ensure CDCMeta for table %s: %w", tableName, err)
		}

		// once we create SelfCDCMeta, read table schema and insert in __cdc table, the schema of the table as
		// fist record
		var schemaSQL string
		err = db.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&schemaSQL)
		if err != nil {
			return 0, fmt.Errorf("failed to get schema for table %s: %w", tableName, err)
		}

		h := sha1.New()
		h.Write([]byte(schemaSQL))
		hash := hex.EncodeToString(h.Sum(nil))

		// now update current_schema_hash in SelfCDCMeta
		_, err = db.Exec(`
			UPDATE SelfCDCMeta 
			SET current_schema_hash = ? 
			WHERE table_name = ?
		`, hash, tableName)
		if err != nil {
			return 0, fmt.Errorf("failed to update current_schema_hash for table %s: %w", tableName, err)
		}

		newEntries++
	}

	return newEntries, nil
}

// ensureCDCMeta ensures a SelfCDCMeta record exists for the given table.
// If the table already has records, it sets cdc_start_id to the max primary key value.
func ensureCDCMeta(db *sql.DB, tableName string, pkColumn string) error {
	// Check if CDCMeta record already exists
	var existingID int
	err := db.QueryRow("SELECT id FROM SelfCDCMeta WHERE table_name = ?", tableName).Scan(&existingID)
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
			UPDATE SelfCDCMeta 
			SET cdc_start_id = ? 
			WHERE table_name = ? AND cdc_start_id = 0
		`, cdcStartID, tableName)
		if err != nil {
			return fmt.Errorf("failed to update SelfCDCMeta for table %s: %w", tableName, err)
		}
	} else {
		// Insert new record
		_, err = db.Exec(`
			INSERT INTO SelfCDCMeta (table_name, cdc_start_id, current_cdc_id, gc_max_records, last_gc_at, extrameta)
			VALUES (?, ?, 0, 0, 0, '{}')
		`, tableName, cdcStartID)
		if err != nil {
			return fmt.Errorf("failed to insert SelfCDCMeta for table %s: %w", tableName, err)
		}

		// once we create SelfCDCMeta, read table schema and insert in __cdc table, the schema of the table as
		// fist record
		var schemaSQL string
		err = db.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&schemaSQL)
		if err != nil {
			return fmt.Errorf("failed to get schema for table %s: %w", tableName, err)
		}

		cdcTableName := tableName + "__cdc"

		_, err = db.Exec(fmt.Sprintf("INSERT INTO %s (record_id, operation, schema_text) VALUES (0, 3, ?)", cdcTableName), schemaSQL)
		if err != nil {
			return fmt.Errorf("failed to insert schema init record for %s: %w", tableName, err)
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
		if strings.HasSuffix(tableName, "__cdc") {
			continue
		}

		// drop triggers on the table
		if _, err := db.Exec("DROP TRIGGER IF EXISTS ?", tableName+"_insert__cdc"); err != nil {
			return fmt.Errorf("failed to drop INSERT trigger for %s: %w", tableName, err)
		}
		if _, err := db.Exec("DROP TRIGGER IF EXISTS ?", tableName+"_update__cdc"); err != nil {
			return fmt.Errorf("failed to drop UPDATE trigger for %s: %w", tableName, err)
		}
		if _, err := db.Exec("DROP TRIGGER IF EXISTS ?", tableName+"_delete__cdc"); err != nil {
			return fmt.Errorf("failed to drop DELETE trigger for %s: %w", tableName, err)
		}

		cdcTableName := tableName + "__cdc"
		if _, err := db.Exec("DROP TABLE IF EXISTS ?", cdcTableName); err != nil {
			return fmt.Errorf("failed to drop CDC table %s: %w", cdcTableName, err)
		}

	}

	// truncate SelfCDCMeta table
	if _, err := db.Exec("TRUNCATE TABLE SelfCDCMeta"); err != nil {
		return fmt.Errorf("failed to truncate SelfCDCMeta table: %w", err)
	}

	return nil

}
