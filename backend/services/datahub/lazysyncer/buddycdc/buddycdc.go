package buddycdc

import (
	"fmt"
	"path/filepath"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazytypes"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

type BuddyCDC struct {
	buddyPubKey string
	mainDb      db.Session
	dbSession   db.Session
	state       map[int64]int64

	transport lazytypes.RemoteBuddyTransport
}

func NewBuddyCDC(maindb db.Session, basePath, buddyPubKey string) (*BuddyCDC, error) {

	dbSession, err := sqlite.Open(sqlite.ConnectionURL{
		Database: filepath.Join(basePath, fmt.Sprintf("buddycdc_%s.db", buddyPubKey)),
	})
	if err != nil {
		return nil, err
	}

	buddyCDC := &BuddyCDC{
		mainDb:      maindb,
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
