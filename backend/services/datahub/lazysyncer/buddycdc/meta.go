package buddycdc

import (
	"fmt"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazymodel"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/upper/db/v4"
)

func (b *BuddyCDC) evLoop() {

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
