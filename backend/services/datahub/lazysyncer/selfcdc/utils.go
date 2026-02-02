package selfcdc

import (
	"database/sql"
	"fmt"
	"strings"
)

// getPrimaryKeyColumn returns the primary key column name for a table.
// It uses PRAGMA table_info to find the primary key column.
// If no explicit primary key is found, it defaults to "rowid" (SQLite's implicit rowid).
func getPrimaryKeyColumn(db *sql.DB, tableName string) (string, error) {
	// Quote table name to handle special characters safely
	quotedTableName := fmt.Sprintf(`"%s"`, strings.ReplaceAll(tableName, `"`, `""`))
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", quotedTableName))
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var pkColumn string
	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue interface{}
		var pk int

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return "", err
		}

		if pk > 0 {
			// Found primary key column
			pkColumn = name
			break
		}
	}

	if err := rows.Err(); err != nil {
		return "", err
	}

	// If no explicit primary key found, use rowid (SQLite's implicit primary key)
	if pkColumn == "" {
		pkColumn = "rowid"
	}

	return pkColumn, nil
}

func getTableNames(db *sql.DB) ([]string, error) {
	// Add query optimization hint for SQLite
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tableNames = append(tableNames, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tableNames, nil
}

// getExistingCDCTables checks existence of multiple CDC tables in one query
func getExistingCDCTables(db *sql.DB, cdcTableNames []string) (map[string]bool, error) {
	if len(cdcTableNames) == 0 {
		return make(map[string]bool), nil
	}

	// Build IN clause with placeholders
	placeholders := make([]string, len(cdcTableNames))
	args := make([]interface{}, len(cdcTableNames))
	for i, name := range cdcTableNames {
		placeholders[i] = "?"
		args[i] = name
	}

	query := fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='table' AND name IN (%s)",
		strings.Join(placeholders, ","))

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	existing := make(map[string]bool)
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		existing[tableName] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return existing, nil
}
