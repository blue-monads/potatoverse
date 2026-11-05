package space

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

func (d *SpaceOperations) AddSpacePlugin(data *dbmodels.SpacePlugin) (int64, error) {
	table := d.spacePluginsTable()
	r, err := table.Insert(data)
	if err != nil {
		return 0, err
	}
	return r.ID().(int64), nil
}

func (d *SpaceOperations) GetSpacePlugin(id int64) (*dbmodels.SpacePlugin, error) {
	table := d.spacePluginsTable()
	data := &dbmodels.SpacePlugin{}
	err := table.Find(db.Cond{"id": id}).One(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *SpaceOperations) ListSpacePlugins(sourceInstallId, sourceSpaceId int64) ([]dbmodels.SpacePlugin, error) {
	table := d.spacePluginsTable()
	datas := make([]dbmodels.SpacePlugin, 0)
	cond := db.Cond{"source_install_id": sourceInstallId}
	if sourceSpaceId != 0 {
		cond["source_space_id"] = sourceSpaceId
	}
	err := table.Find(cond).All(&datas)
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (d *SpaceOperations) QuerySpacePlugins(cond map[any]any) ([]dbmodels.SpacePlugin, error) {
	table := d.spacePluginsTable()
	datas := make([]dbmodels.SpacePlugin, 0)
	err := table.Find(db.Cond(cond)).All(&datas)
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (d *SpaceOperations) UpdateSpacePlugin(id int64, data map[string]any) error {
	table := d.spacePluginsTable()
	return table.Find(db.Cond{"id": id}).Update(data)
}

func (d *SpaceOperations) RemoveSpacePlugin(id int64) error {
	table := d.spacePluginsTable()
	return table.Find(db.Cond{"id": id}).Delete()
}

func (d *SpaceOperations) ListSpacesBySpaceType(spaceType string) ([]dbmodels.Space, error) {
	datas := make([]dbmodels.Space, 0)
	err := d.spaceTable().Find(db.Cond{"space_type": spaceType}).All(&datas)
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (d *SpaceOperations) spacePluginsTable() db.Collection {
	return d.db.Collection("SpacePlugins")
}
