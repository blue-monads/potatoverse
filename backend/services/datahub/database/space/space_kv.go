package space

import (
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

func (d *SpaceOperations) QuerySpaceKV(installId int64, cond map[any]any, offset int, limit int) ([]dbmodels.SpaceKV, error) {
	table := d.spaceKVTable()
	datas := make([]dbmodels.SpaceKV, 0)

	cond["install_id"] = installId

	if limit > 1000 || limit <= 0 {
		limit = 100
	}

	err := table.Find(db.Cond(cond)).
		Select("id", "key", "group", "tag1", "tag2", "tag3").
		OrderBy("id ASC").
		Offset(offset).
		Limit(limit).
		All(&datas)
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (d *SpaceOperations) AddSpaceKV(spaceId int64, data *dbmodels.SpaceKV) error {
	table := d.spaceKVTable()
	_, err := table.Insert(data)
	if err != nil {
		return err
	}
	return nil
}

func (d *SpaceOperations) GetSpaceKVByID(spaceId int64, id int64) (*dbmodels.SpaceKV, error) {
	table := d.spaceKVTable()
	data := &dbmodels.SpaceKV{}
	err := table.Find(db.Cond{"install_id": spaceId, "id": id}).One(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *SpaceOperations) GetSpaceKV(spaceId int64, group string, key string) (*dbmodels.SpaceKV, error) {
	table := d.spaceKVTable()
	data := &dbmodels.SpaceKV{}
	err := table.Find(db.Cond{"install_id": spaceId, "group": group, "key": key}).One(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *SpaceOperations) GetSpaceKVByGroup(spaceId int64, group string, offset int, limit int) ([]dbmodels.SpaceKV, error) {
	table := d.spaceKVTable()
	datas := make([]dbmodels.SpaceKV, 0)
	err := table.Find(db.Cond{"install_id": spaceId, "group": group}).Offset(offset).Limit(limit).All(&datas)
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (d *SpaceOperations) UpdateSpaceKV(spaceId int64, group, key string, data map[string]any) error {
	table := d.spaceKVTable()
	return table.Find(db.Cond{"install_id": spaceId, "group": group, "key": key}).Update(data)
}

func (d *SpaceOperations) UpsertSpaceKV(spaceId int64, group, key string, data map[string]any) error {
	table := d.spaceKVTable()
	cond := db.Cond{"install_id": spaceId, "group": group, "key": key}

	exists, err := table.Find(cond).Exists()
	if err != nil {
		return err
	}

	if exists {
		delete(data, "install_id")
		return table.Find(cond).Update(data)
	}

	data["install_id"] = spaceId
	data["group"] = group
	data["key"] = key

	_, err = table.Insert(data)
	if err != nil {
		return err
	}

	return nil

}

func (d *SpaceOperations) RemoveSpaceKV(spaceId int64, group string, key string) error {
	table := d.spaceKVTable()
	return table.Find(db.Cond{"install_id": spaceId, "group": group, "key": key}).Delete()
}

func (d *SpaceOperations) spaceKVTable() db.Collection {
	return d.db.Collection("SpaceKV")
}
