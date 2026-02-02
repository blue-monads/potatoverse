package selfcdc2

import (
	"database/sql"
	"slices"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazymodel"
)

func (s *SelfCDCSyncer) UnapplyCDC() error {

	tables, err := s.getTableNames()
	if err != nil {
		return err
	}

	sqlconn := s.db.Driver().(*sql.DB)

	for _, tableName := range tables {
		if strings.HasSuffix(tableName, "__cdc") {
			continue
		}

		if slices.Contains(lazymodel.SkipTables, tableName) {
			continue
		}

		dropStmt, err := buildDropTriggerSchema(tableName)
		if err != nil {
			return err
		}

		_, err = sqlconn.Exec(dropStmt)
		if err != nil {
			return err
		}

	}

	return nil
}
