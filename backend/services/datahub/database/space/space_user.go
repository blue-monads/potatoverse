package space

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

func (d *SpaceOperations) QuerySpaceUsers(installId int64, cond map[any]any) ([]dbmodels.SpaceUser, error) {
	table := d.spaceUserTable()
	datas := make([]dbmodels.SpaceUser, 0)

	cond["install_id"] = installId

	err := table.Find(db.Cond(cond)).All(&datas)
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (d *SpaceOperations) AddSpaceUser(installId int64, data *dbmodels.SpaceUser) (int64, error) {
	data.InstallID = installId
	table := d.spaceUserTable()
	r, err := table.Insert(data)
	if err != nil {
		return 0, err
	}
	return r.ID().(int64), nil
}

func (d *SpaceOperations) GetSpaceUser(installId int64, id int64) (*dbmodels.SpaceUser, error) {
	table := d.spaceUserTable()
	data := &dbmodels.SpaceUser{}
	err := table.Find(db.Cond{"install_id": installId, "id": id}).One(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *SpaceOperations) UpdateSpaceUser(installId int64, id int64, data map[string]any) error {
	table := d.spaceUserTable()
	return table.Find(db.Cond{"install_id": installId, "id": id}).Update(data)
}

func (d *SpaceOperations) RemoveSpaceUser(installId int64, id int64) error {
	table := d.spaceUserTable()
	return table.Find(db.Cond{"install_id": installId, "id": id}).Delete()
}
