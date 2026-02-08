package space

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

func (d *SpaceOperations) QueryAllEventSubscriptions(includeDisabled bool) ([]dbmodels.MQSubscriptionLite, error) {
	table := d.eventSubscriptionTable()
	datas := make([]dbmodels.MQSubscriptionLite, 0)
	cond := db.Cond{}
	if !includeDisabled {
		cond["disabled"] = false
	}
	err := table.Find(cond).All(&datas)
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (d *SpaceOperations) QueryEventSubscriptions(installId int64, cond map[any]any) ([]dbmodels.MQSubscription, error) {
	table := d.eventSubscriptionTable()
	datas := make([]dbmodels.MQSubscription, 0)

	cond["install_id"] = installId

	err := table.Find(db.Cond(cond)).All(&datas)
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (d *SpaceOperations) AddEventSubscription(installId int64, data *dbmodels.MQSubscription) (int64, error) {
	data.InstallID = installId
	table := d.eventSubscriptionTable()
	r, err := table.Insert(data)
	if err != nil {
		return 0, err
	}
	return r.ID().(int64), nil
}

func (d *SpaceOperations) GetEventSubscription(installId int64, id int64) (*dbmodels.MQSubscription, error) {
	table := d.eventSubscriptionTable()
	data := &dbmodels.MQSubscription{}
	err := table.Find(db.Cond{"install_id": installId, "id": id}).One(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *SpaceOperations) UpdateEventSubscription(installId int64, id int64, data map[string]any) error {
	table := d.eventSubscriptionTable()
	return table.Find(db.Cond{"install_id": installId, "id": id}).Update(data)
}

func (d *SpaceOperations) RemoveEventSubscription(installId int64, id int64) error {
	table := d.eventSubscriptionTable()
	return table.Find(db.Cond{"install_id": installId, "id": id}).Delete()
}

func (d *SpaceOperations) eventSubscriptionTable() db.Collection {
	return d.db.Collection("MQSubscriptions")
}
