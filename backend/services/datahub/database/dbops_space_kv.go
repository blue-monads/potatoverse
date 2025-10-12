package database

import (
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

func (d *DB) QuerySpaceKV(spaceId int64, cond map[any]any) ([]dbmodels.SpaceKV, error) {
	table := d.spaceKVTable()
	datas := make([]dbmodels.SpaceKV, 0)

	cond["space_id"] = spaceId

	err := table.Find(db.Cond(cond)).All(&datas)
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (d *DB) AddSpaceKV(spaceId int64, data *dbmodels.SpaceKV) error {
	table := d.spaceKVTable()
	_, err := table.Insert(data)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) GetSpaceKV(spaceId int64, group string, key string) (*dbmodels.SpaceKV, error) {
	table := d.spaceKVTable()
	data := &dbmodels.SpaceKV{}
	err := table.Find(db.Cond{"space_id": spaceId, "group": group, "key": key}).One(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *DB) GetSpaceKVByGroup(spaceId int64, group string, offset int, limit int) ([]dbmodels.SpaceKV, error) {
	table := d.spaceKVTable()
	datas := make([]dbmodels.SpaceKV, 0)
	err := table.Find(db.Cond{"space_id": spaceId, "group": group}).Offset(offset).Limit(limit).All(&datas)
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (d *DB) UpdateSpaceKV(spaceId int64, group, key string, data map[string]any) error {
	table := d.spaceKVTable()
	return table.Find(db.Cond{"space_id": spaceId, "group": group, "key": key}).Update(data)
}

func (d *DB) UpsertSpaceKV(spaceId int64, group, key string, data map[string]any) error {
	table := d.spaceKVTable()
	cond := db.Cond{"space_id": spaceId, "group": group, "key": key}

	exists, err := table.Find(cond).Exists()
	if err != nil {
		return err
	}

	if exists {
		delete(data, "space_id")
		return table.Find(cond).Update(data)
	}

	data["space_id"] = spaceId
	data["group"] = group
	data["key"] = key

	_, err = table.Insert(data)
	if err != nil {
		return err
	}

	return nil

}

func (d *DB) RemoveSpaceKV(spaceId int64, group string, key string) error {
	table := d.spaceKVTable()
	return table.Find(db.Cond{"space_id": spaceId, "group": group, "key": key}).Delete()
}

func (d *DB) spaceKVTable() db.Collection {
	return d.Table("SpaceKV")
}
