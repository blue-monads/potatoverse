package selfcdc

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazytypes"
	"github.com/upper/db/v4"
)

func (c *SelfCDCSyncer) ApplyCDC() error {

	tableNames, err := c.getTableNames()
	if err != nil {
		return err
	}

	for _, tableName := range tableNames {

		if strings.HasSuffix(tableName, "__cdc") {
			continue
		}

		if slices.Contains(lazytypes.SkipTables, tableName) {
			continue
		}

		exists, err := c.tableExists(tableName + "__cdc")
		if err != nil {
			return err
		}

		if exists {
			continue
		}

		if err := c.ensureCDC(tableName); err != nil {
			return err
		}

	}

	return nil

}

func (c *SelfCDCSyncer) ensureCDC(tableName string) error {
	cdcTable := tableName + "__cdc"

	pkColumn, err := c.getPrimaryKeyColumn(tableName)
	if err != nil {
		return fmt.Errorf("failed to get primary key for table %s: %w", tableName, err)
	}

	cdcTableSQL, err := lazytypes.BuildSelfCDCTableSchema(cdcTable)
	if err != nil {
		return fmt.Errorf("failed to build template for table %s: %w", tableName, err)
	}

	sqlconn := c.db.Driver().(*sql.DB)

	_, err = sqlconn.Exec(cdcTableSQL)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create CDC table %s: %w", cdcTable, err)
		}
	}

	triggerSQL, err := buildTriggerSchema(tableName, pkColumn)
	if err != nil {
		return fmt.Errorf("failed to build template for table %s: %w", tableName, err)
	}

	_, err = sqlconn.Exec(triggerSQL)
	if err != nil {
		return fmt.Errorf("failed to create trigger for table %s: %w", tableName, err)
	}

	tinfo, err := c.getTableInfo(tableName)
	if err != nil {
		return fmt.Errorf("failed to get table info for table %s: %w", tableName, err)
	}

	shash := hashTableSchema(tinfo.Sql)

	meta, err := c.GetCDCMeta(tableName)
	if err != nil {
		if err == db.ErrNoMoreRows {
			// create meta

			maxRowid, err := c.tableMaxId(tableName, pkColumn)
			if err != nil {
				return fmt.Errorf("failed to get max rowid for table %s: %w", tableName, err)
			}

			_, err = c.selfcdcTable().Insert(map[string]any{
				"table_name":          tableName,
				"current_schema_hash": "",
				"start_row_id":        maxRowid,
				"primary_key":         pkColumn,
			})
			if err != nil {
				return fmt.Errorf("failed to insert cdc meta for table %s: %w", tableName, err)
			}
			meta, err = c.GetCDCMeta(tableName)
			if err != nil {
				return fmt.Errorf("failed to get cdc meta for table %s after insert: %w", tableName, err)
			}
		} else {
			return fmt.Errorf("failed to get cdc meta for table %s: %w", tableName, err)
		}
	}

	if meta.CurrentSchemaHash != shash {

		isInit := meta.CurrentSchemaHash == ""

		err := c.setHash(tableName, tinfo.Sql, shash, isInit)
		if err != nil {
			return fmt.Errorf("failed to set schema for table %s: %w", tableName, err)
		}

	}

	return nil
}

func (s *SelfCDCSyncer) setHash(tableName string, schema, shash string, isInit bool) error {

	// 3: schema_init 4:schema_change
	opId := 4
	if isInit {
		opId = 3
	}

	_, err := s.db.Collection(tableName + "__cdc").Insert(map[string]any{
		"record_id": 0,
		"operation": opId,
		"payload":   []byte(schema),
	})

	if err != nil {
		return fmt.Errorf("failed to insert schema for table %s: %w", tableName, err)
	}

	err = s.selfcdcTable().Find(db.Cond{"table_name": tableName}).Update(map[string]any{"current_schema_hash": shash})
	if err != nil {
		return fmt.Errorf("failed to update schema hash for table %s: %w", tableName, err)
	}

	return nil

}

func (s *SelfCDCSyncer) GetCDCMeta(tableName string) (*lazytypes.SelfCDCMeta, error) {
	var cdcMeta lazytypes.SelfCDCMeta
	err := s.selfcdcTable().Find(db.Cond{"table_name": tableName}).One(&cdcMeta)
	if err != nil {
		return nil, err
	}

	return &cdcMeta, nil
}

// private

func (c *SelfCDCSyncer) tableExists(tableName string) (bool, error) {
	row, err := c.db.SQL().QueryRow("SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", tableName)
	if err != nil {
		return false, err
	}

	var name string
	err = row.Scan(&name)
	if err != nil {
		return false, nil // Row not found
	}

	return true, nil
}

func (s *SelfCDCSyncer) selfcdcTable() db.Collection {
	return s.db.Collection("SelfCDCMeta")
}

func hashTableSchema(schema string) string {

	h := sha256.Sum256([]byte(schema))

	return hex.EncodeToString(h[:])
}
