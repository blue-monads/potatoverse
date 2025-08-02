package buddy

import (
	"fmt"

	"github.com/upper/db/v4"
)

type DCC struct {
	Id     int64
	RowId  int64
	TypeId uint8
	Data   []byte
}

type BuddyStore struct {
	db                 db.Session
	buddyId            string
	maxStorageSize     int
	currentStorageSize int
}

type ChangeModel struct {
	ID      int64  `json:"id" db:"id,omitempty"`
	LastSid int32  `json:"last_sid" db:"last_sid"`
	RowId   int64  `json:"row_id" db:"row_id"`
	TypeId  uint8  `json:"type_id" db:"type_id"`
	Data    []byte `json:"data" db:"data"`
}

func (b *BuddyStore) CleanupOldRows(lastSid int64, table string) error {
	return b.tableCollection(table).Find(db.Cond{
		"last_sid <": lastSid,
	}).Delete()
}

func (b *BuddyStore) WriteBatchedDDC(lastSid int32, tableId string, data []DCC) error {
	tableName := b.tableName(tableId)
	tbl := b.db.Collection(tableName)

	exists, err := tbl.Exists()
	if err != nil {
		return err
	}

	if !exists {
		err = b.AddBuddyTable(tableName)
		if err != nil {
			return err
		}
	}

	finalData := make([]ChangeModel, 0, len(data))

	for _, d := range data {
		finalData = append(finalData, ChangeModel{
			ID:      d.Id,
			LastSid: lastSid,
			RowId:   d.RowId,
			TypeId:  d.TypeId,
			Data:    d.Data,
		})
	}

	batcher := b.db.SQL().InsertInto(tableName).Columns(
		"id", "row_id", "last_sid", "type_id", "data",
	).Batch(len(data))

	for _, d := range finalData {
		batcher.Values(d)
	}

	batcher.Done()

	err = batcher.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (b *BuddyStore) AddBuddyTable(fullTableName string) error {

	schema := fmt.Sprintf(`
	CREATE IF NOT EXISTS %s(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		row_id INTEGER NOT NULL,
		last_sid INTEGER NOT NULL,
		type_id INTEGER NOT NULL,
		data BLOB
	);`, fullTableName)

	_, err := b.db.SQL().Exec(schema)
	if err != nil {
		return err
	}

	return nil

}

var buddyInline = `

CREATE TABLE IF NOT EXISTS BuddyInlineSnapshots (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  buddy_id TEXT NOT NULL,
  blob BLOB
);

`

func (b *BuddyStore) ensureBuddyTable() error {
	_, err := b.db.SQL().Exec(buddyInline)
	if err != nil {
		return err
	}

	return nil

}

// private

func (b *BuddyStore) tableName(name string) string {
	tableName := fmt.Sprintf("__xBuddy_%s_%s", b.buddyId, name)
	return tableName
}

func (b *BuddyStore) tableCollection(name string) db.Collection {
	return b.db.Collection(b.tableName(name))
}
