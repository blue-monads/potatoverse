package lazytypes

var SkipTables = []string{
	"SelfCDCMeta",
	"BuddyCDCMeta",
}

type SelfCDCMeta struct {
	Id                int64  `json:"id" db:"id,omitempty"`
	TableName         string `json:"table_name" db:"table_name"`
	PrimaryKey        string `json:"primary_key" db:"primary_key"`
	StartRowID        int64  `json:"start_row_id" db:"start_row_id"`
	MaxRowID          int64  `json:"max_row_id" db:"-"`
	CurrentMaxCDCID   int64  `json:"current_max_cdc_id" db:"current_max_cdc_id"`
	CurrentCDCID      int64  `json:"current_cdc_id" db:"current_cdc_id"`
	GCMaxRecords      int64  `json:"gc_max_records" db:"gc_max_records"`
	LastGCAt          int64  `json:"last_gc_at" db:"last_gc_at"`
	LastCachedAt      string `json:"last_cached_at" db:"last_current_cached_at"`
	CurrentSchemaHash string `json:"current_schema_hash" db:"current_schema_hash"`
}
