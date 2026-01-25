package cdc

import (
	"time"

	"github.com/upper/db/v4"
)

type CDCMeta struct {
	TableName    string    `json:"table_name"`
	CDCStartID   int64     `json:"cdc_start_id"`
	CurrentCDCID int64     `json:"current_cdc_id"`
	GCMaxRecords int64     `json:"gc_max_records"`
	LastGCAt     time.Time `json:"last_gc_at"`
}

func (s *CDCSyncer) UpdateCurrentCdcId(tableName string) error {
	// query table for max rowid
	row, err := s.db.SQL().QueryRow("SELECT MAX(rowid) FROM ?", tableName)
	if err != nil {
		return err
	}

	var maxRowid int64
	if err := row.Scan(&maxRowid); err != nil {
		return err
	}

	newData := map[string]any{"current_cdc_id": maxRowid}

	// update current_cdc_id in CDCMeta table
	return s.tableName().Find(db.Cond{"table_name": tableName}).Update(newData)
}

func (s *CDCSyncer) tableName() db.Collection {
	return s.db.Collection("CDCMeta")
}
