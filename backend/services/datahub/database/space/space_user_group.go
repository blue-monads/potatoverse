package space

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

func (d *SpaceOperations) QuerySpaceUserGroups(installId int64, cond map[any]any) ([]dbmodels.SpaceUserGroup, error) {
	table := d.spaceUserGroupTable()
	datas := make([]dbmodels.SpaceUserGroup, 0)

	if installId != 0 {
		cond["install_id"] = installId
	}

	err := table.Find(db.Cond(cond)).All(&datas)
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (d *SpaceOperations) AddSpaceUserGroup(installId int64, data *dbmodels.SpaceUserGroup) (int64, error) {
	data.InstallID = installId
	table := d.spaceUserGroupTable()
	r, err := table.Insert(data)
	if err != nil {
		return 0, err
	}
	return r.ID().(int64), nil
}

func (d *SpaceOperations) GetSpaceUserGroup(installId int64, id int64) (*dbmodels.SpaceUserGroup, error) {
	table := d.spaceUserGroupTable()
	data := &dbmodels.SpaceUserGroup{}
	err := table.Find(db.Cond{"install_id": installId, "id": id}).One(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *SpaceOperations) UpdateSpaceUserGroup(installId int64, id int64, data map[string]any) error {
	table := d.spaceUserGroupTable()
	return table.Find(db.Cond{"install_id": installId, "id": id}).Update(data)
}

func (d *SpaceOperations) RemoveSpaceUserGroup(installId int64, id int64) error {
	table := d.spaceUserGroupTable()
	return table.Find(db.Cond{"install_id": installId, "id": id}).Delete()
}
