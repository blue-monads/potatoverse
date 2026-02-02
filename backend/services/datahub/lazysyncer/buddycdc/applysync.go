package buddycdc

import (
	"time"

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

				if err := b.updateMeta(localMeta.Id, map[string]any{
					"start_row_id": data.SyncTillId,
				}); err != nil {
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
							// localMeta.SyncedCDCID = data.SyncTillId
							// b.updateMeta(localMeta

							err = b.updateMeta(localMeta.Id, map[string]any{
								"synced_cdc_id": data.SyncTillId,
							})
							if err != nil {
								qq.Println("@err updating", err.Error())
							}

							continue
						}
						break
					}

					if err := b.saveRecords(localMeta.TableName, data.Records); err != nil {
						qq.Println("Error saving CDC records:", err)
						break
					}
					qq.Println("Synced CDC records:", len(data.Records), "for table:", localMeta.TableName)

					err = b.updateMeta(localMeta.Id, map[string]any{
						"synced_cdc_id": data.SyncTillId,
					})
					if err != nil {
						qq.Println("@err updating", err.Error())
					}

				}
			}
		}

	}

}

func (b *BuddyCDC) updateMeta(tid int64, data map[string]any) error {
	btable := b.buddyMetaTable()
	return btable.Find(db.Cond{"id": tid}).Update(data)
}
