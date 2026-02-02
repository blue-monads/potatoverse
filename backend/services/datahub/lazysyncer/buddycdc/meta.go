package buddycdc

import (
	"errors"
	"fmt"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazymodel"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/upper/db/v4"
)

func (b *BuddyCDC) evLoop() {

	for {

		time.Sleep(10 * time.Second)

		currTables, err := b.getMetaForTables()
		if err != nil {
			qq.Println("Error fetching buddy meta from local:", err)
			continue
		}

		if len(currTables) == 0 {
			tables, err := b.transport.GetMeta()
			if err != nil {
				qq.Println("Error fetching buddy meta:", err)
				continue
			}

			err = b.applyTablesMeta(tables)
			if err != nil {
				qq.Println("Error applying buddy meta:", err)
				continue
			}
		}

		tables, err := b.transport.GetMeta()
		if err != nil {
			qq.Println("Error fetching buddy meta:", err)
			continue
		}

		for _, remoteTableMeta := range tables {
			localMeta, err := b.getMetaForTableId(remoteTableMeta.RemoteTableID)
			if err != nil {
				qq.Println("Error fetching buddy meta for table id:", remoteTableMeta.RemoteTableID, err)
				continue
			}

			// 1. Sync Serial Data First (Historical data before CDC was enabled)
			for localMeta.SyncedRowID < localMeta.StartRowID {
				data, err := b.transport.GetDataSerial(localMeta.RemoteTableID, localMeta.SyncedRowID)
				if err != nil {
					qq.Println("Error fetching serial data:", err)
					break
				}
				if data == nil || len(data.Records) == 0 {
					break
				}

				if err := b.saveRecords(localMeta.TableName, data.Records); err != nil {
					qq.Println("Error saving serial records:", err)
					break
				}
				qq.Println("Synced serial records:", len(data.Records), "for table:", localMeta.TableName)

				localMeta.SyncedRowID = data.SyncTillId
				if err := b.updateMeta(localMeta); err != nil {
					qq.Println("Error updating meta after serial sync:", err)
					break
				}
			}

			// 2. Sync CDC Data (Incremental updates)
			// Only start CDC sync if we are caught up with serial sync (or if there was no serial sync needed)
			if localMeta.SyncedRowID >= localMeta.StartRowID {
				for localMeta.SyncedCDCID < remoteTableMeta.CurrentMaxCDCID {
					data, err := b.transport.GetDataCDC(localMeta.RemoteTableID, localMeta.SyncedCDCID)
					if err != nil {
						qq.Println("Error fetching CDC data:", err)
						break
					}
					if data == nil || len(data.Records) == 0 {
						// If we got no records but the remote says it's ahead, maybe we just advance or wait?
						// For now let's break to avoid infinite loop if something is wrong
						// OR rely on SyncTillId from response if it advances even empty
						if data != nil && data.SyncTillId > localMeta.SyncedCDCID {
							localMeta.SyncedCDCID = data.SyncTillId
							b.updateMeta(localMeta)
							continue
						}
						break
					}

					if err := b.saveRecords(localMeta.TableName, data.Records); err != nil {
						qq.Println("Error saving CDC records:", err)
						break
					}
					qq.Println("Synced CDC records:", len(data.Records), "for table:", localMeta.TableName)

					localMeta.SyncedCDCID = data.SyncTillId
					if err := b.updateMeta(localMeta); err != nil {
						qq.Println("Error updating meta after CDC sync:", err)
						break
					}
				}
			}
		}

	}

}

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

func (b *BuddyCDC) updateMeta(meta *lazymodel.BuddyCDCMeta) error {
	btable := b.buddyMetaTable()
	return btable.Find(db.Cond{"id": meta.Id}).Update(meta)
}

func (b *BuddyCDC) applyTablesMeta(tables []*lazymodel.BuddyCDCMeta) error {

	for _, tableMeta := range tables {
		remoteTableId := tableMeta.RemoteTableID

		_, err := b.getMetaForTableId(remoteTableId)
		if err != nil {
			if errors.Is(err, db.ErrNoMoreRows) {

				tableName := buddyTable(remoteTableId)
				qq.Println("Creating new buddy table:", tableName, "for remote table:", tableMeta.TableName)

				// 1. Create Meta
				newMeta := &lazymodel.BuddyCDCMeta{
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

				cdcTableSQL, err := lazymodel.BuildCDCTableSchema(tableName)
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

func (b *BuddyCDC) getMetaForTableId(tableId int64) (*lazymodel.BuddyCDCMeta, error) {
	meta := &lazymodel.BuddyCDCMeta{}
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

func (b *BuddyCDC) getMetaForTables() ([]*lazymodel.BuddyCDCMeta, error) {

	metas := []*lazymodel.BuddyCDCMeta{}

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
