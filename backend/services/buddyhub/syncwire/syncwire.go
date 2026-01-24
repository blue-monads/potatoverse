package syncwire

const (
	SyncWireInitSyncType = iota
	SyncWireSyncDataType = iota
)

type SyncWireInit struct {
	AllowPush bool `json:"allow_push"`
}

type SyncWireInitResponse struct {
	Tables map[string]int64 `json:"tables"`
}

type SyncWireSyncData struct {
	Table          string `json:"table"`
	CurrentCdcHead int64  `json:"current_cdc_head"`
}

type Record struct {
	RowId     int64  `json:"row_id"`
	Operation int    `json:"operation"`
	Data      []byte `json:"data"`
}

type SyncWireSyncDataResponse struct {
	Table   string `json:"table"`
	Records []any  `json:"records"`
}
