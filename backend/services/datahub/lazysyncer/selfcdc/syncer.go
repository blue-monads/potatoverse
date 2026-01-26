package selfcdc

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

type SelfCDCSyncer struct {
	db            db.Session
	isEnabled     bool
	ontableChange chan string

	cdcIdIndex map[int64]string
	stateCache map[string]*CDCMeta
	mu         sync.RWMutex
}

func NewSelfCDCSyncer(db db.Session, isEnabled bool) *SelfCDCSyncer {
	return &SelfCDCSyncer{
		db:            db,
		isEnabled:     isEnabled,
		ontableChange: make(chan string, 100),
		cdcIdIndex:    make(map[int64]string),
		stateCache:    make(map[string]*CDCMeta),
		mu:            sync.RWMutex{},
	}
}

func (s *SelfCDCSyncer) Start() error {
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

func (s *SelfCDCSyncer) updateStateCache() error {
	cmetas, err := s.GetAllCdcMeta()
	if err != nil {
		return err
	}

	cdcIdIndex := make(map[int64]string)
	stateCache := make(map[string]*CDCMeta)

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

func (s *SelfCDCSyncer) pollSyncLoop() {
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

func (s *SelfCDCSyncer) notifySyncLoop() {

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

func (s *SelfCDCSyncer) GetAllCdcMeta() ([]*CDCMeta, error) {
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

func (s *SelfCDCSyncer) GetTableRecords(tableName string, offset int64, limit int64) ([]map[string]any, error) {
	table := s.db.Collection(tableName)
	var records []map[string]any
	err := table.Find(db.Cond{"rowid >": offset}).Limit(int(limit)).All(&records)
	if err != nil {
		return nil, err
	}

	return records, nil
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
