package buddycdc

import (
	"errors"
	"fmt"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazytypes"
	"github.com/upper/db/v4"
)

func (b *BuddyCDC) saveRecords(tableName string, records []lazytypes.Record) error {
	tbl := b.dbSession.Collection(tableName)

	for _, record := range records {

		data := map[string]any{
			"record_id":     record.RecordId,
			"linked_cdc_id": record.LinkedCDCId,
			"operation":     record.Operation,
		}

		if record.Payload != nil {
			data["payload"] = record.Payload
		}

		_, err := tbl.Insert(data)
		if err != nil {
			return err
		}

	}
	return nil
}

func (b *BuddyCDC) applyTablesMeta(tables []*lazytypes.SelfCDCMeta) error {

	for _, tableMeta := range tables {

		_, err := b.getMetaForTableName(tableMeta.TableName)
		if err != nil {
			if errors.Is(err, db.ErrNoMoreRows) {

				// 1. Create Meta
				newMeta := &lazytypes.BuddyCDCMeta{
					PubKey:          b.buddyPubKey,
					RemoteTableID:   tableMeta.Id,
					TableName:       tableMeta.TableName,
					StartRowID:      tableMeta.StartRowID,
					SyncedRowID:     0,
					CurrentMaxCDCID: 0,
					SyncedCDCID:     0,
				}

				_, err := b.buddyMetaTable().Insert(newMeta)
				if err != nil {
					return fmt.Errorf("failed to insert buddy meta: %w", err)
				}

				cdcTableSQL, err := lazytypes.BuildBuddyCDCTableSchema(tableMeta.TableName)
				if err != nil {
					return fmt.Errorf("failed to build template for table %s: %w", tableMeta.TableName, err)
				}

				if _, err := b.dbSession.SQL().Exec(cdcTableSQL); err != nil {
					return fmt.Errorf("failed to create table %s: %w", tableMeta.TableName, err)
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

func (b *BuddyCDC) getMetaForTableName(tableName string) (*lazytypes.BuddyCDCMeta, error) {
	meta := &lazytypes.BuddyCDCMeta{}
	btable := b.buddyMetaTable()

	err := btable.Find(db.Cond{
		"table_name": tableName,
		"pubkey":     b.buddyPubKey,
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
