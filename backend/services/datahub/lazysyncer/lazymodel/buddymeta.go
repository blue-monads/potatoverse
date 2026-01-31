package lazymodel

type BuddyCDCMeta struct {
	Id             int64  `json:"id" db:"id"`
	PubKey         string `json:"pubkey" db:"pubkey"`
	RemoteTableID  int64  `json:"remote_table_id" db:"remote_table_id"`
	TableName      string `json:"table_name" db:"table_name"`
	CDCStartID     int64  `json:"cdc_start_id" db:"cdc_start_id"`
	CurrentCDCID   int64  `json:"current_cdc_id" db:"current_cdc_id"`
	InitSchemaText string `json:"init_schema_text" db:"init_schema_text"`
	IsDeleted      bool   `json:"is_deleted" db:"is_deleted"`
}
