package selfcdc

import (
	"time"

	"github.com/upper/db/v4"
)

type CDCMeta struct {
	TableName    string     `json:"table_name"`
	CDCStartID   int64      `json:"cdc_start_id"`
	CurrentCDCID int64      `json:"current_cdc_id"`
	GCMaxRecords int64      `json:"gc_max_records"`
	LastGCAt     *time.Time `json:"last_gc_at"`
	LastCachedAt *time.Time `json:"last_cached_at"`
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
	err = s.tableName().Find(db.Cond{"table_name": tableName}).Update(newData)
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

func (s *SelfCDCSyncer) GetCDCMeta(tableName string) (*CDCMeta, error) {
	var cdcMeta CDCMeta
	err := s.tableName().Find(db.Cond{"table_name": tableName}).One(&cdcMeta)
	if err != nil {
		return nil, err
	}

	return &cdcMeta, nil
}

func (s *SelfCDCSyncer) tableName() db.Collection {
	return s.db.Collection("CDCMeta")
}
