package selfcdc

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazymodel"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/upper/db/v4"
)

const CACHE_INTERVAL = 30 * time.Second
const NotifyMode = true

type SelfCDCSyncer struct {
	db            db.Session
	isEnabled     bool
	ontableChange chan string

	cdcIdIndex map[int64]string
	stateCache map[string]*lazymodel.SelfCDCMeta
	mu         sync.RWMutex
}

func NewSelfCDCSyncer(db db.Session, isEnabled bool) *SelfCDCSyncer {
	return &SelfCDCSyncer{
		db:            db,
		isEnabled:     isEnabled,
		ontableChange: make(chan string, 100),
		cdcIdIndex:    make(map[int64]string),
		stateCache:    make(map[string]*lazymodel.SelfCDCMeta),
		mu:            sync.RWMutex{},
	}
}

func (s *SelfCDCSyncer) Start() error {

	if s.isEnabled {
		if err := s.ApplyCDC(); err != nil {
			return err
		}
	} else {
		// if err := s.DropCDC(); err != nil {
		// 	return err
		// }
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

func (s *SelfCDCSyncer) updateStateCache() error {
	cmetas, err := s.GetAllCdcMeta()
	if err != nil {
		return err
	}

	cdcIdIndex := make(map[int64]string)
	stateCache := make(map[string]*lazymodel.SelfCDCMeta)

	for _, cmeta := range cmetas {
		cdcIdIndex[cmeta.CurrentCDCID] = cmeta.TableName
		stateCache[cmeta.TableName] = cmeta
	}

	s.mu.Lock()
	s.cdcIdIndex = cdcIdIndex
	s.stateCache = stateCache
	s.mu.Unlock()

	return nil
}

func (s *SelfCDCSyncer) AttachMissingTables() error {
	if !s.isEnabled {
		return nil
	}

	if err := s.ApplyCDC(); err != nil {
		return err
	}

	if err := s.updateStateCache(); err != nil {
		return err
	}

	return nil

}

func (s *SelfCDCSyncer) pollSyncLoop() {
	for {
		time.Sleep(CACHE_INTERVAL)

		alltables, err := s.getTableNames()
		if err != nil {
			continue
		}

		for _, tableName := range alltables {

			if slices.Contains(lazymodel.SkipTables, tableName) {
				continue
			}

			if strings.HasSuffix(tableName, "__cdc") {
				continue
			}

			currentCdcId, err := s.UpdateCurrentCdcId(tableName)
			if err != nil {
				continue
			}

			qq.Println("currentCdcId", currentCdcId)
		}

	}

}

func (s *SelfCDCSyncer) notifySyncLoop() {

	readAllPendingTables := func() []string {
		tables := make([]string, 0, 1)

		timer := time.NewTimer(5)
		defer timer.Stop()

		for {
			select {
			case tableName := <-s.ontableChange:

				if slices.Contains(tables, tableName) {
					continue
				}

				tables = append(tables, tableName)
			case <-timer.C:
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

func (s *SelfCDCSyncer) GetAllCdcMeta() ([]*lazymodel.SelfCDCMeta, error) {

	var cdcMeta []*lazymodel.SelfCDCMeta
	err := s.selfcdcTable().Find().All(&cdcMeta)
	if err != nil {
		return nil, err
	}

	return cdcMeta, nil
}

func (s *SelfCDCSyncer) getTableName(tblId int64) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tableName, ok := s.cdcIdIndex[tblId]
	if !ok {
		return ""
	}

	return tableName
}

var (
	ErrTableNotFound = fmt.Errorf("table not found")
)

func (s *SelfCDCSyncer) GetTableRecordsSerial(tblId int64, offset int64, limit int64) (map[int64]map[string]any, error) {
	tableName := s.getTableName(tblId)
	if tableName == "" {
		return nil, ErrTableNotFound
	}

	table := s.db.Collection(tableName)
	var records []map[string]any
	err := table.Find(db.Cond{"rowid >": offset}).Select("rowid", "*").Limit(int(limit)).All(&records)
	if err != nil {
		return nil, err
	}

	final := make(map[int64]map[string]any, len(records))
	for _, record := range records {
		rowidAny, ok := record["rowid"]
		if !ok {
			continue
		}

		rowid, ok := rowidAny.(int64)
		if !ok {
			continue
		}

		final[rowid] = record
	}

	return final, nil
}

func (s *SelfCDCSyncer) GetTableRecords(tableId int64, ids []int64) (map[int64]map[string]any, error) {
	tableName := s.getTableName(tableId)
	if tableName == "" {
		return nil, ErrTableNotFound
	}

	table := s.db.Collection(tableName)
	var records []map[string]any
	err := table.Find(db.Cond{"rowid": ids}).All(&records)
	if err != nil {
		return nil, err
	}

	final := make(map[int64]map[string]any, len(records))
	for _, record := range records {
		rowidAny, ok := record["rowid"]
		if !ok {
			continue
		}

		rowid, ok := rowidAny.(int64)
		if !ok {
			continue
		}

		final[rowid] = record
	}

	return final, nil
}

func (s *SelfCDCSyncer) GetCDCCache() map[int64]int64 {
	cache := make(map[int64]int64)

	s.mu.RLock()
	defer s.mu.RUnlock()

	for id, tableName := range s.cdcIdIndex {
		cmeta, ok := s.stateCache[tableName]
		if !ok {
			continue
		}

		cache[id] = cmeta.CurrentCDCID
	}

	return cache
}

func (s *SelfCDCSyncer) UpdateCurrentCdcId(tableName string) (int64, error) {
	// query table for max rowid
	row, err := s.db.SQL().QueryRow("SELECT MAX(rowid) FROM ?", tableName)
	if err != nil {
		return 0, err
	}

	var maxRowid int64
	if err := row.Scan(&maxRowid); err != nil {
		return 0, err
	}

	newData := map[string]any{"current_cdc_id": maxRowid}

	// update current_cdc_id in CDCMeta table
	err = s.selfcdcTable().Find(db.Cond{"table_name": tableName}).Update(newData)
	if err != nil {
		return 0, err
	}

	cmeta, err := s.GetCDCMeta(tableName)
	if err != nil {
		return 0, err
	}

	s.mu.Lock()
	s.stateCache[tableName] = cmeta
	s.mu.Unlock()

	return maxRowid, nil
}
