package selfcdc

import (
	"fmt"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazytypes"
	"github.com/upper/db/v4"
)

var (
	ErrTableNotFound = fmt.Errorf("table not found")
)

func (s *SelfCDCSyncer) GetMeta() ([]*lazytypes.SelfCDCMeta, error) {
	metas, err := s.GetAllCdcMeta()
	if err != nil {
		return nil, err
	}
	return metas, nil
}
func (s *SelfCDCSyncer) GetDataSerial(tableId int64, sinceRowId int64) (*lazytypes.BuddyData, error) {

	records, err := s.GetTableRecordsSerial(tableId, sinceRowId, 100)
	if err != nil {
		return nil, err
	}

	maxRowId := int64(0)
	for rowid := range records {
		if rowid > maxRowId {
			maxRowId = rowid
		}
	}

	return &lazytypes.BuddyData{
		Records:       records,
		TableCDCIndex: map[int64]int64{tableId: maxRowId},
		SyncTillId:    maxRowId,
	}, nil
}

func (s *SelfCDCSyncer) GetDataCDC(tableId int64, sinceCdcId int64) (*lazytypes.BuddyData, error) {

	// list cdc since cdc id
	// make list of rows to fetch
	// fetch records

	return nil, nil
}

//

func (s *SelfCDCSyncer) GetTableRecordsSerial(tblId int64, offset int64, limit int64) (map[int64]map[string]any, error) {
	tableName := s.getTableName(tblId)
	if tableName == "" {
		return nil, ErrTableNotFound
	}

	table := s.db.Collection(tableName)
	var records []map[string]any
	err := table.Find(db.Cond{"rowid >": offset}).Select("rowid", "*").Limit(int(limit)).All(&records)
	if err != nil {
		return nil, err
	}

	final := make(map[int64]map[string]any, len(records))
	for _, record := range records {
		rowidAny, ok := record["rowid"]
		if !ok {
			continue
		}

		rowid, ok := rowidAny.(int64)
		if !ok {
			continue
		}

		final[rowid] = record
	}

	return final, nil
}

func (s *SelfCDCSyncer) GetTableRecords(tableId int64, ids []int64) (map[int64]map[string]any, error) {
	tableName := s.getTableName(tableId)
	if tableName == "" {
		return nil, ErrTableNotFound
	}

	table := s.db.Collection(tableName)
	var records []map[string]any
	err := table.Find(db.Cond{"rowid": ids}).All(&records)
	if err != nil {
		return nil, err
	}

	final := make(map[int64]map[string]any, len(records))
	for _, record := range records {
		rowidAny, ok := record["rowid"]
		if !ok {
			continue
		}

		rowid, ok := rowidAny.(int64)
		if !ok {
			continue
		}

		final[rowid] = record
	}

	return final, nil
}
