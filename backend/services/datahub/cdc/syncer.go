package cdc

import (
	"database/sql"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/upper/db/v4"
)

const CACHE_INTERVAL = 30 * time.Second
const NotifyMode = true

type CDCSyncer struct {
	db            db.Session
	isEnabled     bool
	ontableChange chan string

	cdcIdIndex map[int64]string
	stateCache map[string]*CDCMeta
	mu         sync.RWMutex
}

func NewCDCSyncer(db db.Session, isEnabled bool) *CDCSyncer {
	return &CDCSyncer{
		db:            db,
		isEnabled:     isEnabled,
		ontableChange: make(chan string, 100),
		cdcIdIndex:    make(map[int64]string),
		stateCache:    make(map[string]*CDCMeta),
		mu:            sync.RWMutex{},
	}
}

func (s *CDCSyncer) Start() error {
	driver := s.db.Driver().(*sql.DB)

	if s.isEnabled {
		if _, err := EnsureCDC(driver); err != nil {
			return err
		}
	} else {
		if err := DropCDC(driver); err != nil {
			return err
		}
	}

	if err := s.updateStateCache(); err != nil {
		return err
	}

	if NotifyMode {
		go s.notifySyncLoop()
	} else {
		go s.pollSyncLoop()
	}

	return nil
}

func (s *CDCSyncer) updateStateCache() error {
	cmetas, err := s.GetAllCdcMeta()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cdcIdIndex = make(map[int64]string)
	s.stateCache = make(map[string]*CDCMeta)

	for _, cmeta := range cmetas {
		s.cdcIdIndex[cmeta.CurrentCDCID] = cmeta.TableName
		s.stateCache[cmeta.TableName] = cmeta
	}

	return nil
}

func (s *CDCSyncer) AttachMissingTables() error {
	if !s.isEnabled {
		return nil
	}

	driver := s.db.Driver().(*sql.DB)

	newEntries, err := EnsureCDC(driver)
	if err != nil {
		return err
	}

	if newEntries > 0 {
		if err := s.updateStateCache(); err != nil {
			return err
		}
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

			currentCdcId, err := s.UpdateCurrentCdcId(table)
			if err != nil {
				continue
			}

			qq.Println("currentCdcId", currentCdcId)
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
			_, err := s.UpdateCurrentCdcId(tableName)
			if err != nil {
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
