package lazytypes

type BuddyCDCMeta struct {
	Id            int64  `json:"id" db:"id,omitempty"`
	PubKey        string `json:"pubkey" db:"pubkey"`
	RemoteTableID int64  `json:"remote_table_id" db:"remote_table_id"`
	TableName     string `json:"table_name" db:"table_name"`

	StartRowID  int64 `json:"start_row_id" db:"start_row_id"`
	SyncedRowID int64 `json:"synced_row_id" db:"synced_row_id"`

	CurrentMaxCDCID int64 `json:"current_max_cdc_id" db:"current_max_cdc_id"`
	SyncedCDCID     int64 `json:"synced_cdc_id" db:"synced_cdc_id"`

	CurrentSchemaHash string `json:"current_schema_hash" db:"current_schema_hash"`
	IsDeleted         bool   `json:"is_deleted" db:"is_deleted"`
}
