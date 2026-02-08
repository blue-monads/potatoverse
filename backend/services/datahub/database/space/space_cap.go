package space

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

func (d *SpaceOperations) QuerySpaceCapabilities(installId int64, cond map[any]any) ([]dbmodels.SpaceCapability, error) {
	table := d.spaceCapabilitiesTable()
	datas := make([]dbmodels.SpaceCapability, 0)

	cond["install_id"] = installId

	err := table.Find(db.Cond(cond)).All(&datas)
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (d *SpaceOperations) AddSpaceCapability(installId int64, data *dbmodels.SpaceCapability) error {
	data.InstallID = installId
	table := d.spaceCapabilitiesTable()
	_, err := table.Insert(data)
	if err != nil {
		return err
	}
	return nil
}

func (d *SpaceOperations) GetSpaceCapability(installId int64, name string) (*dbmodels.SpaceCapability, error) {
	table := d.spaceCapabilitiesTable()
	data := &dbmodels.SpaceCapability{}
	err := table.Find(db.Cond{"install_id": installId, "name": name}).One(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *SpaceOperations) GetSpaceCapabilityByID(installId int64, id int64) (*dbmodels.SpaceCapability, error) {
	table := d.spaceCapabilitiesTable()
	data := &dbmodels.SpaceCapability{}
	err := table.Find(db.Cond{"install_id": installId, "id": id}).One(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *SpaceOperations) UpdateSpaceCapability(installId int64, name string, data map[string]any) error {
	table := d.spaceCapabilitiesTable()
	return table.Find(db.Cond{"install_id": installId, "name": name}).Update(data)
}

func (d *SpaceOperations) UpdateSpaceCapabilityByID(installId int64, id int64, data map[string]any) error {
	table := d.spaceCapabilitiesTable()
	return table.Find(db.Cond{"install_id": installId, "id": id}).Update(data)
}

func (d *SpaceOperations) RemoveSpaceCapability(installId int64, name string) error {
	table := d.spaceCapabilitiesTable()
	return table.Find(db.Cond{"install_id": installId, "name": name}).Delete()
}

func (d *SpaceOperations) RemoveSpaceCapabilityByID(installId int64, id int64) error {
	table := d.spaceCapabilitiesTable()
	return table.Find(db.Cond{"install_id": installId, "id": id}).Delete()
}

func (d *SpaceOperations) spaceCapabilitiesTable() db.Collection {
	return d.db.Collection("SpaceCapabilities")
}
