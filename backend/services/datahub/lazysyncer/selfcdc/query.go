package selfcdc

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazytypes"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
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

	datas, err := s.GetTableRecordsSerial(tableId, sinceRowId, 100)
	if err != nil {
		return nil, err
	}

	records := make([]lazytypes.Record, 0, len(datas))

	maxRowId := int64(sinceRowId)
	for _, data := range datas {
		rowidAny, ok := data["rowid"]
		if !ok {
			rowidAny, ok = data["id"]
		}

		if !ok {
			continue
		}

		rowid, ok := rowidAny.(int64)
		if !ok {
			continue
		}

		if rowid > maxRowId {
			maxRowId = rowid
		}

		payload, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		records = append(records, lazytypes.Record{
			RecordId:    rowid,
			Operation:   0,
			LinkedCDCId: 0,
			Payload:     payload,
		})

	}

	qq.Println("@sending_with_max_rowid", maxRowId)

	return &lazytypes.BuddyData{
		Records:       records,
		TableCDCIndex: map[int64]int64{tableId: maxRowId},
		SyncTillId:    maxRowId,
	}, nil
}

type cdcRow struct {
	Id        int64  `db:"id"`
	RecordId  int64  `db:"record_id"`
	Operation int64  `db:"operation"`
	Payload   []byte `db:"payload"`
}

func (s *SelfCDCSyncer) GetDataCDC(tableId int64, sinceCdcId int64) (*lazytypes.BuddyData, error) {
	tableName := s.getTableName(tableId)
	if tableName == "" {
		return nil, ErrTableNotFound
	}

	cdcTable := tableName + "__cdc"

	var cdcRows []cdcRow

	// fetch limit 100
	err := s.db.Collection(cdcTable).Find(db.Cond{"id >": sinceCdcId}).Limit(100).OrderBy("id").All(&cdcRows)
	if err != nil {
		return nil, err
	}

	if len(cdcRows) == 0 {
		return &lazytypes.BuddyData{
			Records:       nil,
			TableCDCIndex: map[int64]int64{tableId: sinceCdcId},
			SyncTillId:    sinceCdcId,
		}, nil
	}

	// rowid => index
	uniqueRecordIds := make(map[int64]int)
	recordIds := make([]int64, 0, len(cdcRows))
	maxCdcId := sinceCdcId
	records := make([]lazytypes.Record, 0)

	for idx, cdcRow := range cdcRows {

		if cdcRow.Operation == 3 || cdcRow.Operation == 4 {

			records = append(records, lazytypes.Record{
				RecordId:    cdcRow.RecordId,
				Operation:   cdcRow.Operation,
				LinkedCDCId: cdcRow.Id,
				Payload:     cdcRow.Payload,
			})

			if cdcRow.Id > maxCdcId {
				maxCdcId = cdcRow.Id
			}

			continue
		}

		existingEntry, ok := uniqueRecordIds[cdcRow.RecordId]
		if ok {
			if cdcRows[existingEntry].Id > cdcRow.Id {
				continue
			}
		}

		uniqueRecordIds[cdcRow.RecordId] = idx
		recordIds = append(recordIds, cdcRow.RecordId)

		if cdcRow.Id > maxCdcId {
			maxCdcId = cdcRow.Id
		}
	}

	datas, err := s.GetTableRecords(tableId, recordIds)
	if err != nil {
		return nil, err
	}

	for _, data := range datas {
		rowidAny, ok := data["rowid"]
		if !ok {
			rowidAny, ok = data["id"]
		}

		if !ok {
			continue
		}

		rowid, ok := rowidAny.(int64)
		if !ok {
			continue
		}

		cdcRow := cdcRows[uniqueRecordIds[rowid]]

		payload, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		records = append(records, lazytypes.Record{
			RecordId:    rowid,
			Operation:   cdcRow.Operation,
			LinkedCDCId: cdcRow.Id,
			Payload:     payload,
		})
	}

	slices.SortFunc(records, func(a, b lazytypes.Record) int {
		return int(a.LinkedCDCId - b.LinkedCDCId)
	})

	return &lazytypes.BuddyData{
		Records:       records,
		TableCDCIndex: map[int64]int64{tableId: maxCdcId},
		SyncTillId:    maxCdcId,
	}, nil
}

//

func (s *SelfCDCSyncer) GetTableRecordsSerial(tblId int64, offset int64, limit int64) ([]map[string]any, error) {
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

	return records, nil
}

func (s *SelfCDCSyncer) GetTableRecords(tableId int64, ids []int64) ([]map[string]any, error) {
	tableName := s.getTableName(tableId)
	if tableName == "" {
		return nil, ErrTableNotFound
	}

	table := s.db.Collection(tableName)
	var records []map[string]any
	err := table.Find(db.Cond{"rowid": ids}).Select("rowid", "*").All(&records)
	if err != nil {
		return nil, err
	}

	return records, nil
}
