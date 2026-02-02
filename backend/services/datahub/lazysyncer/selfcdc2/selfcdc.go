package selfcdc2

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazymodel"
	"github.com/upper/db/v4"
)

type CDCMaker struct {
	db db.Session
}

func NewCDCMaker(db db.Session) *CDCMaker {
	return &CDCMaker{db: db}
}

func (c *CDCMaker) ApplyCDC() error {

	tableNames, err := c.getTableNames()
	if err != nil {
		return err
	}

	for _, tableName := range tableNames {

		if strings.HasSuffix(tableName, "__cdc") {
			continue
		}

		if slices.Contains(lazymodel.SkipTables, tableName) {
			continue
		}

		exists, err := c.tableExists(tableName + "__cdc")
		if err != nil {
			return err
		}

		if exists {
			continue
		}

		if err := c.createCDC(tableName); err != nil {
			return err
		}

	}

	return nil

}

func (c *CDCMaker) createCDC(tableName string) error {
	cdcTable := tableName + "__cdc"

	pkColumn, err := c.getPrimaryKeyColumn(tableName)
	if err != nil {
		return fmt.Errorf("failed to get primary key for table %s: %w", tableName, err)
	}

	cdcTableSQL, err := buildCDCTableSchema(tableName)
	if err != nil {
		return fmt.Errorf("failed to build template for table %s: %w", tableName, err)
	}

	sqlconn := c.db.Driver().(*sql.DB)

	_, err = sqlconn.Exec(cdcTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create CDC table %s: %w", cdcTable, err)
	}

	triggerSQL, err := buildTriggerSchema(tableName, pkColumn)
	if err != nil {
		return fmt.Errorf("failed to build template for table %s: %w", tableName, err)
	}

	_, err = sqlconn.Exec(triggerSQL)
	if err != nil {
		return fmt.Errorf("failed to create trigger for table %s: %w", tableName, err)
	}

	return nil
}

// private

func (c *CDCMaker) tableExists(tableName string) (bool, error) {
	return c.db.Collection(tableName).Exists()
}
