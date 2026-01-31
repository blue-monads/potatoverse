package lazymodel

import "time"

type SelfCDCMeta struct {
	TableName    string     `json:"table_name"`
	CDCStartID   int64      `json:"cdc_start_id"`
	CurrentCDCID int64      `json:"current_cdc_id"`
	GCMaxRecords int64      `json:"gc_max_records"`
	LastGCAt     *time.Time `json:"last_gc_at"`
	LastCachedAt *time.Time `json:"last_cached_at"`
}
