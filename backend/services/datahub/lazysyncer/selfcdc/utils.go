package selfcdc

import (
	"fmt"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

func (c *SelfCDCSyncer) getTableNames() ([]string, error) {

	tableNames := []string{}

	err := c.db.Collection("sqlite_master").Find().All(&tableNames)
	if err != nil {
		return nil, err
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

	err := c.db.Collection("sqlite_master").
		Find(db.Cond{"name": tableName}).One(info)
	if err != nil {
		return nil, err
	}

	return info, nil
}
