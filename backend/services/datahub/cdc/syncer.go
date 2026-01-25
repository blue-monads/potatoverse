package cdc

import (
	"database/sql"
	"strings"
	"time"

	"github.com/upper/db/v4"
)

type CDCSyncer struct {
	db        db.Session
	isEnabled bool
}

func NewCDCSyncer(db db.Session, isEnabled bool) *CDCSyncer {
	return &CDCSyncer{db: db, isEnabled: isEnabled}
}

func (s *CDCSyncer) Start() error {
	driver := s.db.Driver().(*sql.DB)

	if s.isEnabled {
		if err := EnsureCDC(driver); err != nil {
			return err
		}
	} else {
		if err := DropCDC(driver); err != nil {
			return err
		}
	}

	go s.syncLoop()

	return nil
}

func (s *CDCSyncer) AttachMissingTables() error {
	if !s.isEnabled {
		return nil
	}

	driver := s.db.Driver().(*sql.DB)

	if err := EnsureCDC(driver); err != nil {
		return err
	}

	return nil

}

const CACHE_INTERVAL = 30 * time.Second

func (s *CDCSyncer) syncLoop() {
	driver := s.db.Driver().(*sql.DB)

	for {
		time.Sleep(CACHE_INTERVAL)

		alltables, err := getTableNames(driver)
		if err != nil {
			continue
		}

		for _, table := range alltables {
			if table == "CDCMeta" {
				continue
			}

			if strings.HasSuffix(table, "__cdc") {
				continue
			}

			if err := s.UpdateCurrentCdcId(table); err != nil {
				continue
			}
		}

	}

}

func (s *CDCSyncer) GetAllCdcMeta() ([]*CDCMeta, error) {
	table := s.tableName()
	var cdcMeta []*CDCMeta
	err := table.Find().All(&cdcMeta)
	if err != nil {
		return nil, err
	}

	/*

		for _, cdc := range cdcMeta {
			// fixme => scramble / encrypt table name
			cdc.TableName = cdc.TableName

		}

	*/

	return cdcMeta, nil
}

func (s *CDCSyncer) GetTableRecords(tableName string, offset int64, limit int64) ([]map[string]any, error) {
	table := s.db.Collection(tableName)
	var records []map[string]any
	err := table.Find(db.Cond{"rowid >": offset}).Limit(int(limit)).All(&records)
	if err != nil {
		return nil, err
	}

	return records, nil
}
