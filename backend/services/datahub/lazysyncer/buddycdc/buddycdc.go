package buddycdc

import (
	"fmt"
	"path/filepath"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazymodel"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
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
	dbSession   db.Session
	state       map[int64]int64

	transport RemoteBuddyTransport
}

func NewBuddyCDC(basePath, buddyPubKey string) (*BuddyCDC, error) {

	dbSession, err := sqlite.Open(sqlite.ConnectionURL{
		Database: filepath.Join(basePath, fmt.Sprintf("buddycdc_%s.db", buddyPubKey)),
	})
	if err != nil {
		return nil, err
	}

	buddyCDC := &BuddyCDC{
		buddyPubKey: buddyPubKey,
		dbSession:   dbSession,
		state:       make(map[int64]int64),
	}

	buddyCDC.Start()

	return buddyCDC, nil
}

func (b *BuddyCDC) Start() {

	go b.evLoop()
}
