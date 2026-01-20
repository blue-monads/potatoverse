package lazysyncer

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync/atomic"
	"time"

	"github.com/tidwall/wal"
)

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

	// table__<table_name + row_id> -> wal_index -> [32][8][8]
	// wal__<wal_index> -> <table_name + row_id> -> [8][32][8]
}

type Entry struct {
	Table   string
	Id      int64
	GroupOf int64
}

func (e *LazySyncEngine) Notify(table string, id int64) error {
	data, err := json.Marshal(Entry{Table: table, Id: id, GroupOf: 0})
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	e.wal.Write(e.counter.Load(), data)
	return nil
}

func (e *LazySyncEngine) BatchNotify(table string, ids []int64) error {

	groupOf := int64(len(ids))

	for _, id := range ids {

		data, err := json.Marshal(Entry{Table: table, Id: id, GroupOf: groupOf})
		if err != nil {
			return fmt.Errorf("json.Marshal: %w", err)
		}
		e.wal.Write(e.counter.Load(), data)

	}

	return nil

}

func (e *LazySyncEngine) watchWal() error {

	// lastTruncate := uint64(0)
	lastProcessed := uint64(0)

	for {

		time.Sleep(1 * time.Second)

		lastIndex, err := e.wal.LastIndex()
		if err != nil {
			return fmt.Errorf("wal.LastIndex: %w", err)
		}

		if lastIndex <= lastProcessed {
			time.Sleep(2 * time.Second)
			continue
		}

		i := lastProcessed + 1
		for i < lastIndex {
			data, err := e.wal.Read(i)
			if err != nil {
				return fmt.Errorf("wal.Read: %w", err)
			}
			var entry Entry
			err = json.Unmarshal(data, &entry)
			if err != nil {
				return fmt.Errorf("json.Unmarshal: %w", err)
			}
			err = e.processWal(&entry)
			if err != nil {
				return fmt.Errorf("processWal: %w", err)
			}

			lastProcessed = i

		}

	}

}

func (e *LazySyncEngine) processWal(entry *Entry) error {

	return nil
}
