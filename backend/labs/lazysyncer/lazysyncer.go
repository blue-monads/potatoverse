package lazysyncer

import (
	"os"
	"path"

	"github.com/tidwall/buntdb"
	"github.com/tidwall/wal"
)

func NewLazySyncEngine(folder string) (*LazySyncEngine, error) {

	os.MkdirAll(folder, 0755)

	wal, err := wal.Open(path.Join(folder, "lazy_syncer.wal"), nil)
	if err != nil {
		return nil, err
	}

	buntdb, err := buntdb.Open(path.Join(folder, "lazy_syncer.db"))
	if err != nil {
		return nil, err
	}

	return &LazySyncEngine{wal: wal, buntdb: buntdb}, nil

}

type LazySyncEngine struct {
	wal    *wal.Log
	buntdb *buntdb.DB
}
