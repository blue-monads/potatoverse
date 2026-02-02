package lazytypes

type BuddyData struct {
	Records       map[int64]map[string]any `json:"records"`
	TableCDCIndex map[int64]int64          `json:"table_cdc_index"`
	SyncTillId    int64                    `json:"sync_till_id"`
}

type RemoteBuddyTransport interface {
	GetMeta() ([]*BuddyCDCMeta, error)
	GetDataSerial(tableId int64, sinceRowId int64) (*BuddyData, error)
	GetDataCDC(tableId int64, sinceCdcId int64) (*BuddyData, error)
}
