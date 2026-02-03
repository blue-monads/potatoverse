package buddycdc

import (
	"time"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/upper/db/v4"
)

func (b *BuddyCDC) evLoop() {

	for {

		time.Sleep(time.Second * 30)

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

				qq.Println("@start_poll_table_stat", remoteTableMeta.TableName)

				localMeta, err := b.getMetaForTableId(remoteTableMeta.Id)
				if err != nil {
					continue
				}

				// 1. Sync Serial Data First (Historical data before CDC was enabled)
				for localMeta.SyncedRowID < localMeta.StartRowID {
					data, err := b.transport.GetDataSerial(localMeta.RemoteTableID, localMeta.SyncedRowID)
					if err != nil {
						break
					}
					if data == nil || len(data.Records) == 0 {
						break
					}

					if err := b.saveRecords(localMeta.TableName, data.Records); err != nil {
						break
					}

					if err := b.updateMeta(localMeta.Id, map[string]any{
						"start_row_id": data.SyncTillId,
					}); err != nil {
						break
					}
				}

				// 2. Sync CDC Data (Incremental updates)
				// Only start CDC sync if we are caught up with serial sync (or if there was no serial sync needed)
				if localMeta.SyncedRowID >= localMeta.StartRowID {
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

						if err := b.saveRecords(localMeta.TableName, data.Records); err != nil {
							break
						}

						err = b.updateMeta(localMeta.Id, map[string]any{
							"synced_cdc_id": data.SyncTillId,
						})
						if err == nil {
							localMeta.SyncedCDCID = data.SyncTillId
						}
					}
				}

				qq.Println("@end_poll_table_stat", remoteTableMeta.TableName)
			}
		}
	}
}

func (b *BuddyCDC) updateMeta(tid int64, data map[string]any) error {
	btable := b.buddyMetaTable()
	return btable.Find(db.Cond{"id": tid}).Update(data)
}
