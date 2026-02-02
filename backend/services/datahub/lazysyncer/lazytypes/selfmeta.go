package lazytypes

import "time"

var SkipTables = []string{
	"SelfCDCMeta",
	"BuddyCDCMeta",
}

type SelfCDCMeta struct {
	TableName         string     `json:"table_name" db:"table_name"`
	StartRowID        int64      `json:"start_row_id" db:"start_row_id"`
	CurrentMaxCDCID   int64      `json:"current_max_cdc_id" db:"current_max_cdc_id"`
	CurrentCDCID      int64      `json:"current_cdc_id" db:"current_cdc_id"`
	GCMaxRecords      int64      `json:"gc_max_records" db:"gc_max_records"`
	LastGCAt          *time.Time `json:"last_gc_at" db:"last_gc_at"`
	LastCachedAt      *time.Time `json:"last_cached_at" db:"last_cached_at"`
	CurrentSchemaHash string     `json:"current_schema_hash" db:"current_schema_hash"`
}
