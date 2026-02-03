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
		state:       make(map[int64]int64),
		transport:   opts.Transport,
	}

	if err := buddyCDC.ensureSchema(); err != nil {
		return nil, err
	}

	buddyCDC.Start()

	return buddyCDC, nil
}

func (b *BuddyCDC) ensureSchema() error {
	schema := `CREATE TABLE IF NOT EXISTS BuddyCDCMeta (
		id INTEGER PRIMARY KEY,
		pubkey TEXT NOT NULL,
		remote_table_id INTEGER NOT NULL,
		table_name TEXT NOT NULL,
		start_row_id INTEGER NOT NULL DEFAULT 0,
		synced_row_id INTEGER NOT NULL DEFAULT 0,
		current_max_cdc_id INTEGER NOT NULL DEFAULT 0,
		synced_cdc_id INTEGER NOT NULL DEFAULT 0,
		current_cdc_id INTEGER NOT NULL DEFAULT 0,
		current_schema_hash TEXT NOT NULL DEFAULT '',
		is_deleted BOOLEAN NOT NULL DEFAULT 0,
		extrameta JSON NOT NULL DEFAULT '{}'
	);`

	_, err := b.dbSession.SQL().Exec(schema)
	return err
}

func (b *BuddyCDC) Start() {

	go b.evLoop()
}
