package lazysyncer

import (
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

func (b *BuddyAdapter) GetDataCDC(tableId int64, sinceCdcId int64) (*lazytypes.BuddyData, error) {
	req, err := http.NewRequest("GET", "/zz/buddy/lazycdc/sync/data", nil)
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

	var data *lazytypes.BuddyData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
