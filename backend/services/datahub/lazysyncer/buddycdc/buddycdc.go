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
	transport   lazytypes.RemoteBuddyTransport
}

type Options struct {
	MainDb      db.Session
	BasePath    string
	BuddyPubKey string
	Transport   lazytypes.RemoteBuddyTransport
}

func NewBuddyCDC(opts Options) (*BuddyCDC, error) {

	dbSession, err := sqlite.Open(sqlite.ConnectionURL{
		Database: filepath.Join(opts.BasePath, fmt.Sprintf("buddycdc_%s.db", opts.BuddyPubKey)),
	})
	if err != nil {
		return nil, err
	}

	buddyCDC := &BuddyCDC{
		mainDb:      opts.MainDb,
		buddyPubKey: opts.BuddyPubKey,
		dbSession:   dbSession,
		transport:   opts.Transport,
	}

	return buddyCDC, nil
}

func (b *BuddyCDC) Start() error {

	go b.evLoop()

	return nil
}
