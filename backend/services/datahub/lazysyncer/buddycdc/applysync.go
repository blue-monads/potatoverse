package buddycdc

import (
	"fmt"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazytypes"
	"github.com/upper/db/v4"
)

func (b *BuddyCDC) evLoop() {

	for {

		time.Sleep(2 * time.Second)

		currTables, err := b.getMetaForTables()
		if err != nil {
			continue
		}

		if len(currTables) == 0 {
			tables, err := b.transport.GetMeta()
			if err != nil {
				continue
			}

			err = b.applyTablesMeta(tables)
			if err != nil {
				continue
			}
		}

		tables, err := b.transport.GetMeta()
		if err != nil {
			continue
		}

		for _, remoteTableMeta := range tables {

			b.logger.Info("@start_poll_table_stat")

			localMeta, err := b.getMetaForTableId(remoteTableMeta.Id)
			if err != nil {
				continue
			}

			// Sync CDC Data (All data now goes through CDC)
			for localMeta.SyncedCDCID < remoteTableMeta.CurrentMaxCDCID {
				data, err := b.transport.GetDataCDC(localMeta.RemoteTableID, localMeta.SyncedCDCID)
				if err != nil {
					break
				}
				if data == nil || len(data.Records) == 0 {
					if data != nil && data.SyncTillId > localMeta.SyncedCDCID {
						err = b.updateMeta(localMeta.Id, map[string]any{
							"synced_cdc_id": data.SyncTillId,
						})
						if err == nil {
							localMeta.SyncedCDCID = data.SyncTillId
						}
						continue
					}
					break
				}

				if err := b.saveRecords(localMeta.Id, data.Records); err != nil {
					break
				}

				err = b.updateMeta(localMeta.Id, map[string]any{
					"synced_cdc_id": data.SyncTillId,
				})
				if err != nil {
					break
				}

				localMeta.SyncedCDCID = data.SyncTillId
			}

			b.logger.Info("@end_poll_table_stat", "table_name", remoteTableMeta.TableName)
		}
	}

}

func (b *BuddyCDC) saveRecords(tableId int64, records []lazytypes.Record) error {
	tbl := b.dbSession.Collection(fmt.Sprintf("zz_B_%d", tableId))

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

func (b *BuddyCDC) updateMeta(tid int64, data map[string]any) error {
	btable := b.buddyMetaTable()
	return btable.Find(db.Cond{"id": tid}).Update(data)
}
