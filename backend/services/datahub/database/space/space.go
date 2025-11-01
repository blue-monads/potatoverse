package space

import (
	"fmt"
	"log"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

type SpaceOperations struct {
	db db.Session
}

func NewSpaceOperations(db db.Session) *SpaceOperations {
	return &SpaceOperations{
		db: db,
	}
}

func (d *SpaceOperations) AddSpace(data *dbmodels.Space) (int64, error) {
	r, err := d.spaceTable().Insert(data)
	if err != nil {
		return 0, err
	}

	return r.ID().(int64), nil
}

func (d *SpaceOperations) GetSpace(id int64) (*dbmodels.Space, error) {
	data := &dbmodels.Space{}

	err := d.spaceTable().Find(db.Cond{"id": id}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *SpaceOperations) ListSpaces() ([]dbmodels.Space, error) {
	datas := make([]dbmodels.Space, 0)

	err := d.spaceTable().Find().All(&datas)
	if err != nil {
		return nil, err
	}

	return datas, nil
}

func (d *SpaceOperations) UpdateSpace(id int64, data map[string]any) error {
	return d.spaceTable().Find(db.Cond{"id": id}).Update(data)
}

func (d *SpaceOperations) RemoveSpace(id int64) error {
	return d.spaceTable().Find(db.Cond{"id": id}).Delete()
}

func (d *SpaceOperations) ListSpaceUsers(spaceId int64) ([]dbmodels.SpaceUser, error) {

	datas := make([]dbmodels.SpaceUser, 0)

	err := d.spaceUserTable().Find(db.Cond{"space_id": spaceId}).All(&datas)
	if err != nil {
		return nil, err
	}

	return datas, nil
}

func (d *SpaceOperations) AddUserToSpace(ownerId int64, userId int64, spaceId int64) error {

	if !d.isOwner(ownerId, spaceId) {
		return fmt.Errorf("user does not have required scope")
	}

	_, err := d.spaceUserTable().Insert(dbmodels.SpaceUser{
		UserID:  userId,
		SpaceID: spaceId,
	})
	return err
}

func (d *SpaceOperations) RemoveUserFromSpace(ownerId int64, userId int64, spaceId int64) error {
	return d.spaceUserTable().Find(db.Cond{"owner_id": ownerId, "user_id": userId, "space_id": spaceId}).Delete()
}
func (d *SpaceOperations) GetSpaceUserScope(userId int64, spaceId int64) (string, error) {

	data := &dbmodels.SpaceUser{}
	err := d.spaceUserTable().Find(db.Cond{"user_id": userId, "space_id": spaceId}).One(data)
	if err != nil {
		return "", err
	}

	return data.Scope, nil
}

func (d *SpaceOperations) ListOwnSpaces(ownerId int64, spaceType string) ([]dbmodels.Space, error) {
	cond := db.Cond{"owned_by": ownerId}
	if spaceType != "" {
		cond["stype"] = spaceType
	}

	datas := make([]dbmodels.Space, 0)
	err := d.spaceTable().Find(cond).All(&datas)
	if err != nil {
		return nil, err
	}

	return datas, nil
}

type TSpace struct {
	SpaceId int64 `json:"space_id" db:"space_id"`
}

func (d *SpaceOperations) ListThirdPartySpaces(userId int64, spaceType string) ([]dbmodels.Space, error) {
	cond := db.Cond{
		"userId": userId,
	}

	projs := make([]TSpace, 0)

	if spaceType != "" {
		cond["stype"] = spaceType
	}

	err := d.spaceUserTable().Find(cond).Select("space_id").All(&projs)

	if err != nil {
		return nil, err
	}

	projIds := make([]int64, 0, len(projs))

	for _, p := range projs {
		projIds = append(projIds, p.SpaceId)
	}

	datas := make([]dbmodels.Space, 0, len(projs))

	err = d.spaceTable().Find(db.Cond{
		"userId": userId,
		"id IN":  projIds,
	}).All(&datas)

	if err != nil {
		return nil, err
	}

	return datas, nil
}

func (d *SpaceOperations) ListSpacesByPackageId(installedId int64) ([]dbmodels.Space, error) {
	datas := make([]dbmodels.Space, 0)

	err := d.spaceTable().Find(db.Cond{"install_id": installedId}).All(&datas)
	if err != nil {
		return nil, err
	}

	return datas, nil
}

// Space Configs

func (d *SpaceOperations) AddSpaceConfig(spaceId int64, uid int64, data *dbmodels.SpaceConfig) (int64, error) {
	data.SpaceID = spaceId

	rid, err := d.spaceConfigsTable().Insert(data)
	if err != nil {
		return 0, err
	}
	id := rid.ID().(int64)
	return id, nil
}

func (d *SpaceOperations) ListSpaceConfigs(spaceId int64) ([]dbmodels.SpaceConfig, error) {
	configs := make([]dbmodels.SpaceConfig, 0)
	err := d.spaceConfigsTable().Find(db.Cond{"space_id": spaceId}).All(&configs)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func (d *SpaceOperations) GetSpaceConfig(spaceId int64, uid int64, id int64) (*dbmodels.SpaceConfig, error) {
	data := &dbmodels.SpaceConfig{}
	err := d.spaceConfigsTable().Find(db.Cond{"space_id": spaceId, "id": id}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *SpaceOperations) UpdateSpaceConfig(spaceId int64, uid int64, id int64, data map[string]any) error {
	return d.spaceConfigsTable().Find(db.Cond{"space_id": spaceId, "id": id}).Update(&data)
}

func (d *SpaceOperations) RemoveSpaceConfig(spaceId int64, uid int64, id int64) error {
	return d.spaceConfigsTable().Find(db.Cond{"space_id": spaceId, "id": id}).Delete()
}

// private

func (d *SpaceOperations) spaceConfigsTable() db.Collection {
	return d.db.Collection("SpaceConfigs")
}

func (d *SpaceOperations) spaceTable() db.Collection {
	return d.db.Collection("Spaces")
}

func (d *SpaceOperations) spaceUserTable() db.Collection {
	return d.db.Collection("SpaceUsers")
}

func (d *SpaceOperations) isOwner(ownerid int64, spaceId int64) bool {
	exist, err := d.spaceTable().Find(db.Cond{"owned_by": ownerid, "id": spaceId}).Exists()

	if err != nil {
		log.Println("owner check error", err)
		return false
	}

	return exist

}

// Space Capabilities

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
