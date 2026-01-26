package buddycdc

type BuddyCDCMeta struct {
	TableName    string `json:"table_name"`
	CDCStartID   int64  `json:"cdc_start_id"`
	CurrentCDCID int64  `json:"current_cdc_id"`
}

type BuddyCDC struct {
	state map[int64]int64
}
