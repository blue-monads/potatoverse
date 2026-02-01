package buddycdc

import (
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

		for _, tableMeta := range tables {
			localMeta, err := b.getMetaForTableId(tableMeta.RemoteTableID)
			if err != nil {
				qq.Println("Error fetching buddy meta for table id:", tableMeta.RemoteTableID, err)
				continue
			}

			if localMeta.CurrentCDCID < tableMeta.CurrentCDCID {
				qq.Println("Need to sync table:", localMeta.TableName)

			}

			// localMeta.TableName

		}

	}

}

func (b *BuddyCDC) applyTablesMeta(tables []*lazymodel.BuddyCDCMeta) error {

	for _, tableMeta := range tables {
		tableName := tableMeta.TableName

		qq.Println("@table", tableName)

		// meta, err := b.getMetaForTable(tableName)
		// if err != nil {
		// 	if errors.Is(err, db.ErrNoMoreRows) {
		// 		// create new meta
		// 	}

		// 	return err
		// }

	}

	return nil
}

func (b *BuddyCDC) initTableForId(meta *lazymodel.BuddyCDCMeta) error {
	// tableName := tableName(meta.RemoteTableID)

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

func (b *BuddyCDC) getMetaForTable(tableName string) (*lazymodel.BuddyCDCMeta, error) {

	meta := &lazymodel.BuddyCDCMeta{}

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

func (b *BuddyCDC) buddyMetaTable() db.Collection {
	return b.metaDb.Collection("BuddyCDCMeta")
}

func tableName(tableId int64) string {
	return fmt.Sprintf("zz_buddy_%d", tableId)
}
