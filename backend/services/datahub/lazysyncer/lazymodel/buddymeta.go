package lazymodel

type BuddyCDCMeta struct {
	Id            int64  `json:"id"`
	RemoteTableID int64  `json:"remote_table_id"`
	TableName     string `json:"table_name"`
	CDCStartID    int64  `json:"cdc_start_id"`
	CurrentCDCID  int64  `json:"current_cdc_id"`
}
