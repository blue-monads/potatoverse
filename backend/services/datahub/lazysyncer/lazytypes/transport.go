package lazytypes

type BuddyData struct {
	Records       []Record        `json:"records"`
	TableCDCIndex map[int64]int64 `json:"table_cdc_index"`
	SyncTillId    int64           `json:"sync_till_id"`
}

type Record struct {
	Id          int64  `json:"id"`
	LinkedCDCId int64  `json:"linked_cdc_id"`
	Operation   string `json:"operation"`
	Payload     []byte `json:"payload"`
}

type RemoteBuddyTransport interface {
	GetMeta() ([]*SelfCDCMeta, error)
	GetDataSerial(tableId int64, sinceRowId int64) (*BuddyData, error)
	GetDataCDC(tableId int64, sinceCdcId int64) (*BuddyData, error)
}
