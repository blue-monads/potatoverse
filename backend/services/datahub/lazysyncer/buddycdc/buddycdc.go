package buddycdc

import (
	"database/sql"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazymodel"
	"github.com/upper/db/v4"
)

type BuddyData struct {
	Records       map[int64]map[string]any `json:"records"`
	TableCDCIndex map[int64]int64          `json:"table_cdc_index"`
	SyncTillId    int64                    `json:"sync_till_id"`
}

type RemoteBuddyTransport interface {
	GetMeta() ([]*lazymodel.BuddyCDCMeta, error)
	GetDataSerial(tableId int64, sinceRowId int64) (*BuddyData, error)
	GetDataCDC(tableId int64, sinceCdcId int64) (*BuddyData, error)
}

type BuddyCDC struct {
	buddyPubKey string
	metaDb      db.Session
	db          *sql.Conn
	state       map[int64]int64

	transport RemoteBuddyTransport
}

func NewBuddyCDC(db *sql.Conn) *BuddyCDC {
	return &BuddyCDC{
		db:    db,
		state: make(map[int64]int64),
	}
}
