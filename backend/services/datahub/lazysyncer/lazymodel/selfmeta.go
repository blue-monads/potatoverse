package lazymodel

import "time"

type SelfCDCMeta struct {
	TableName      string     `json:"table_name" db:"table_name"`
	CDCStartID     int64      `json:"cdc_start_id" db:"cdc_start_id"`
	CurrentCDCID   int64      `json:"current_cdc_id" db:"current_cdc_id"`
	GCMaxRecords   int64      `json:"gc_max_records" db:"gc_max_records"`
	LastGCAt       *time.Time `json:"last_gc_at" db:"last_gc_at"`
	LastCachedAt   *time.Time `json:"last_cached_at" db:"last_cached_at"`
	InitSchemaHash string     `json:"init_schema_hash" db:"init_schema_hash"`
	InitSchemaText string     `json:"init_schema_text" db:"init_schema_text"`
}
