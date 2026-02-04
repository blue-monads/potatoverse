package lazytypes

type BuddyData struct {
	Records    []Record `json:"records"`
	SyncTillId int64    `json:"sync_till_id"`
}

type Record struct {
	RecordId    int64  `json:"record_id"` // rowid
	Operation   int64  `json:"operation"`
	LinkedCDCId int64  `json:"linked_cdc_id"`
	Payload     []byte `json:"payload"`
}

type RemoteBuddyTransport interface {
	GetMeta() ([]*SelfCDCMeta, error)
	GetDataSerial(tableId int64, sinceRowId int64) (*BuddyData, error)
	GetDataCDC(tableId int64, sinceCdcId int64) (*BuddyData, error)
}
