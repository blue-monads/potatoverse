package buddycdc

import (
	"errors"
	"fmt"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazytypes"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/upper/db/v4"
)

func (b *BuddyCDC) saveRecords(tableName string, records map[int64]map[string]any) error {
	tbl := b.dbSession.Collection(tableName)

	for id, record := range records {
		record["id"] = id // Ensure ID is present in the record

		exists, err := tbl.Find(db.Cond{"id": id}).Exists()
		if err != nil {
			return err
		}

		if exists {
			err = tbl.Find(db.Cond{"id": id}).Update(record)
		} else {
			_, err = tbl.Insert(record)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BuddyCDC) applyTablesMeta(tables []*lazytypes.BuddyCDCMeta) error {

	for _, tableMeta := range tables {
		remoteTableId := tableMeta.RemoteTableID

		_, err := b.getMetaForTableId(remoteTableId)
		if err != nil {
			if errors.Is(err, db.ErrNoMoreRows) {

				tableName := buddyTable(remoteTableId)
				qq.Println("Creating new buddy table:", tableName, "for remote table:", tableMeta.TableName)

				// 1. Create Meta
				newMeta := &lazytypes.BuddyCDCMeta{
					PubKey:          b.buddyPubKey,
					RemoteTableID:   remoteTableId,
					TableName:       tableName,
					StartRowID:      tableMeta.StartRowID,
					SyncedRowID:     0,
					CurrentMaxCDCID: 0,
					SyncedCDCID:     0,
				}

				_, err := b.buddyMetaTable().Insert(newMeta)
				if err != nil {
					return fmt.Errorf("failed to insert buddy meta: %w", err)
				}

				cdcTableSQL, err := lazytypes.BuildCDCTableSchema(tableName)
				if err != nil {
					return fmt.Errorf("failed to build template for table %s: %w", tableName, err)
				}

				if _, err := b.dbSession.SQL().Exec(cdcTableSQL); err != nil {
					return fmt.Errorf("failed to create table %s: %w", tableName, err)
				}

				continue
			}

			return err
		}

	}

	return nil
}

func (b *BuddyCDC) getMetaForTableId(tableId int64) (*lazytypes.BuddyCDCMeta, error) {
	meta := &lazytypes.BuddyCDCMeta{}
	btable := b.buddyMetaTable()

	err := btable.Find(db.Cond{
		"remote_table_id": tableId,
		"pubkey":          b.buddyPubKey,
	}).One(meta)

	if err != nil {
		return nil, err
	}

	return meta, nil
}

func (b *BuddyCDC) getMetaForTables() ([]*lazytypes.BuddyCDCMeta, error) {

	metas := []*lazytypes.BuddyCDCMeta{}

	btable := b.buddyMetaTable()

	err := btable.Find(db.Cond{
		"pubkey": b.buddyPubKey,
	}).All(&metas)

	if err != nil {
		return nil, err
	}

	return metas, nil
}

func (b *BuddyCDC) buddyMetaTable() db.Collection {
	return b.dbSession.Collection("BuddyCDCMeta")
}

func buddyTable(tableId int64) string {
	return fmt.Sprintf("zz_buddy_%d", tableId)
}
