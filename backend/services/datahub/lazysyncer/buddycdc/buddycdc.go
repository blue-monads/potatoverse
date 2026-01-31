package buddycdc

import (
	"database/sql"

	"github.com/upper/db/v4"
)

type BuddyCDC struct {
	buddyPubKey string
	metaDb      db.Session
	db          *sql.Conn
	state       map[int64]int64
}

func NewBuddyCDC(db *sql.Conn) *BuddyCDC {
	return &BuddyCDC{
		db:    db,
		state: make(map[int64]int64),
	}
}
