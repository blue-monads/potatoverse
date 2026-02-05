package selfcdc

import (
	"fmt"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
)

func (c *SelfCDCSyncer) getTableNames() ([]string, error) {

	qq.Println("@getTableNames/1")

	tableNames := []string{}

	qq.Println("@getTableNames/2")

	rows, err := c.db.SQL().Query("SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return nil, err
	}

	qq.Println("@getTableNames/3")

	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tableNames = append(tableNames, name)
	}

	return tableNames, nil

}

func (c *SelfCDCSyncer) getPrimaryKeyColumn(tableName string) (string, error) {
	quotedTableName := fmt.Sprintf(`"%s"`, strings.ReplaceAll(tableName, `"`, `""`))
	rows, err := c.db.SQL().Query(fmt.Sprintf("PRAGMA table_info(%s)", quotedTableName))
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

func (c *SelfCDCSyncer) getTableInfo(tableName string) (*dbmodels.TableInfo, error) {

	info := &dbmodels.TableInfo{}

	row, err := c.db.SQL().QueryRow("SELECT name, type, sql FROM sqlite_master WHERE type = 'table' AND name = ?", tableName)
	if err != nil {
		return nil, err
	}

	if err := row.Scan(&info.Name, &info.Type, &info.Sql); err != nil {
		return nil, err
	}

	return info, nil
}

/*

SELECT AVG(
    (length(column1) + length(column2) + length(column3))
) AS average_row_size
FROM your_table_name;

*/

func (c *SelfCDCSyncer) getColumns(tableName string) ([]string, error) {
	quotedTableName := fmt.Sprintf(`"%s"`, strings.ReplaceAll(tableName, `"`, `""`))
	rows, err := c.db.SQL().Query(fmt.Sprintf("PRAGMA table_info(%s)", quotedTableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue interface{}
		var pk int

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		columns = append(columns, name)
	}
	return columns, nil
}

func (c *SelfCDCSyncer) GetAverageRowSize(tableName string) (int64, error) {

	columns, err := c.getColumns(tableName)
	if err != nil {
		return 0, err
	}

	if len(columns) == 0 {
		return 0, nil
	}

	var colSumParts []string
	for _, col := range columns {
		quotedCol := fmt.Sprintf(`"%s"`, strings.ReplaceAll(col, `"`, `""`))
		colSumParts = append(colSumParts, fmt.Sprintf("IFNULL(length(%s), 0)", quotedCol))
	}

	query := fmt.Sprintf("SELECT AVG(%s) FROM \"%s\"", strings.Join(colSumParts, " + "), strings.ReplaceAll(tableName, `"`, `""`))

	row, err := c.db.SQL().QueryRow(query)
	if err != nil {
		return 0, err
	}

	var avgSize float64
	if err := row.Scan(&avgSize); err != nil {
		return 0, nil // Likely no rows
	}

	return int64(avgSize), nil
}
