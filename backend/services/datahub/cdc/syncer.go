package cdc

import (
	"database/sql"
	"slices"
	"strings"
	"time"

	"github.com/upper/db/v4"
)

const CACHE_INTERVAL = 30 * time.Second
const NotifyMode = true

type CDCSyncer struct {
	db            db.Session
	isEnabled     bool
	ontableChange chan string
}

func NewCDCSyncer(db db.Session, isEnabled bool) *CDCSyncer {
	return &CDCSyncer{
		db:            db,
		isEnabled:     isEnabled,
		ontableChange: make(chan string, 100),
	}
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

	if NotifyMode {
		go s.notifySyncLoop()
	} else {
		go s.pollSyncLoop()
	}

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

func (s *CDCSyncer) pollSyncLoop() {
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

func (s *CDCSyncer) notifySyncLoop() {

	readAllPendingTables := func() []string {
		tables := make([]string, 0, 1)

		for {
			select {
			case tableName := <-s.ontableChange:

				if slices.Contains(tables, tableName) {
					continue
				}

				tables = append(tables, tableName)
			default:
				return tables
			}
		}

	}

	for {
		tables := readAllPendingTables()
		for _, tableName := range tables {
			s.UpdateCurrentCdcId(tableName)
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
