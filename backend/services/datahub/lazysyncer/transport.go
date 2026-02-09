package lazysyncer

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer/lazytypes"
)

type BuddyAdapter struct {
	parent  *LazySyncer
	buddyId string
}

func NewBuddyAdapter(parent *LazySyncer, buddyId string) *BuddyAdapter {
	return &BuddyAdapter{
		parent:  parent,
		buddyId: buddyId,
	}
}

func (b *BuddyAdapter) GetMeta() ([]*lazytypes.SelfCDCMeta, error) {

	if b.parent.transport == nil {
		return nil, errors.New("transport is not set")
	}

	req, err := http.NewRequest("GET", "/zz/buddy/lazycdc/sync/meta", nil)
	if err != nil {
		return nil, err
	}

	resp, err := b.parent.transport.SendBuddy(b.buddyId, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var meta []*lazytypes.SelfCDCMeta
	err = json.Unmarshal(body, &meta)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

type SyncRequest struct {
	TableId      int64 `json:"table_id"`
	LastSyncedId int64 `json:"last_synced_id"`
	Limit        int64 `json:"limit"`
}

func (b *BuddyAdapter) GetDataCDC(tableId int64, sinceCdcId int64) (*lazytypes.BuddyData, error) {

	body, err := json.Marshal(SyncRequest{
		TableId:      tableId,
		LastSyncedId: sinceCdcId,
		Limit:        100,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "/zz/buddy/lazycdc/sync/data", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	resp, err := b.parent.transport.SendBuddy(b.buddyId, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data *lazytypes.BuddyData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
