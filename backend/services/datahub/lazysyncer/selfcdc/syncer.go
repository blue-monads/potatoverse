package selfcdc

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazytypes"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/upper/db/v4"
)

const CACHE_INTERVAL = 2 * time.Second
const NotifyMode = false

type SelfCDCSyncer struct {
	db            db.Session
	isEnabled     bool
	ontableChange chan string
}

func NewSelfCDCSyncer(db db.Session, isEnabled bool) *SelfCDCSyncer {
	return &SelfCDCSyncer{
		db:            db,
		isEnabled:     isEnabled,
		ontableChange: make(chan string, 100),
	}
}

func (s *SelfCDCSyncer) Start() error {

	if s.isEnabled {
		if err := s.ApplyCDC(); err != nil {
			return err
		}
	} else {
		if err := s.UnApplyCDC(); err != nil {
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

func (s *SelfCDCSyncer) AttachMissingTables() error {
	if !s.isEnabled {
		return nil
	}

	if err := s.ApplyCDC(); err != nil {
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

			qq.Println("@start_poll_table_stat", tableName)

			if slices.Contains(lazytypes.SkipTables, tableName) {
				continue
			}

			if strings.HasSuffix(tableName, "__cdc") {
				continue
			}

			currentCdcId, err := s.UpdateCurrentCdcId(tableName)
			if err != nil {
				continue
			}

			qq.Println("@end_poll_table_stat", tableName, currentCdcId)

		}

	}

}

func (s *SelfCDCSyncer) notifySyncLoop() {

	readAllPendingTables := func() []string {
		tables := make([]string, 0, 1)

		timer := time.NewTimer(time.Second * 5)
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

func (s *SelfCDCSyncer) GetAllCdcMeta() ([]*lazytypes.SelfCDCMeta, error) {
	var cdcMeta []*lazytypes.SelfCDCMeta
	err := s.selfcdcTable().Find().All(&cdcMeta)
	if err != nil {
		return nil, err
	}

	return cdcMeta, nil
}

func (s *SelfCDCSyncer) tableMaxId(tableName, idColumn string) (int64, error) {
	row, err := s.db.SQL().QueryRow(fmt.Sprintf("SELECT MAX(%s) FROM %s", idColumn, tableName))
	if err != nil {
		return 0, err
	}

	var maxRowid int64
	if err := row.Scan(&maxRowid); err != nil {
		// If no rows, maxRowid will be nil/0, Scan into int64 works for 0
		return 0, nil
	}

	return maxRowid, nil
}

func (s *SelfCDCSyncer) UpdateCurrentCdcId(tableName string) (int64, error) {

	maxRowid, err := s.tableMaxId(tableName+"__log", "id")
	if err != nil {
		return 0, err
	}

	newData := map[string]any{
		"current_max_cdc_id": maxRowid,
	}

	err = s.selfcdcTable().Find(db.Cond{"table_name": tableName}).Update(newData)
	if err != nil {
		return 0, err
	}

	return maxRowid, nil
}
