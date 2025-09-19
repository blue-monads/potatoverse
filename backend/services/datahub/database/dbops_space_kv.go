package database

import (
	"github.com/blue-monads/turnix/backend/services/datahub/models"
	"github.com/upper/db/v4"
)

func (d *DB) QuerySpaceKV(spaceId int64, cond map[any]any) ([]models.SpaceKV, error) {
	table := d.spaceKVTable()
	datas := make([]models.SpaceKV, 0)

	cond["space_id"] = spaceId

	err := table.Find(db.Cond(cond)).All(&datas)
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (d *DB) AddSpaceKV(spaceId int64, data *models.SpaceKV) error {
	table := d.spaceKVTable()
	_, err := table.Insert(data)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) GetSpaceKV(spaceId int64, group string, key string) (*models.SpaceKV, error) {
	table := d.spaceKVTable()
	data := &models.SpaceKV{}
	err := table.Find(db.Cond{"space_id": spaceId, "group": group, "key": key}).One(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *DB) GetSpaceKVByGroup(spaceId int64, group string, offset int, limit int) ([]models.SpaceKV, error) {
	table := d.spaceKVTable()
	datas := make([]models.SpaceKV, 0)
	err := table.Find(db.Cond{"space_id": spaceId, "group": group}).Offset(offset).Limit(limit).All(&datas)
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (d *DB) UpdateSpaceKV(spaceId int64, group string, key string, value string) error {
	table := d.spaceKVTable()
	return table.Find(db.Cond{"space_id": spaceId, "group": group, "key": key}).Update(map[string]any{"value": value})
}

func (d *DB) RemoveSpaceKV(spaceId int64, group string, key string) error {
	table := d.spaceKVTable()
	return table.Find(db.Cond{"space_id": spaceId, "group": group, "key": key}).Delete()
}

func (d *DB) spaceKVTable() db.Collection {
	return d.Table("SpaceKV")
}
