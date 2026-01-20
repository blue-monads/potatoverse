package lazysyncer

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync/atomic"

	"github.com/tidwall/wal"
)

type BuddyEntry struct {
	Table string
	RowId int64
	Mode  int // insert, update, delete, insert_future_delete, update_future_delete
}

func NewLazySyncEngine(folder string) (*LazySyncEngine, error) {

	os.MkdirAll(folder, 0755)

	wal, err := wal.Open(path.Join(folder, "lazy_syncer.wal"), nil)
	if err != nil {
		return nil, err
	}

	return &LazySyncEngine{wal: wal}, nil

}

type LazySyncEngine struct {
	counter atomic.Uint64
	wal     *wal.Log

	lruCache any

	// table__<table_name + row_id> -> wal_index -> [32][8][8]
	// wal__<wal_index> -> <table_name + row_id> -> [8][32][8]
}

type Entry struct {
	Table string
	Id    int64
	Mode  int64 // 0: insert, 1: update, 2: delete
}

func (e *LazySyncEngine) Notify(table string, id int64) error {
	data, err := json.Marshal(Entry{Table: table, Id: id, Mode: 0})
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	e.wal.Write(e.counter.Load(), data)
	return nil
}

func (e *LazySyncEngine) BatchNotify(table string, ids []int64) error {

	for _, id := range ids {

		data, err := json.Marshal(Entry{Table: table, Id: id, Mode: 0})
		if err != nil {
			return fmt.Errorf("json.Marshal: %w", err)
		}
		e.wal.Write(e.counter.Load(), data)

	}

	return nil

}
