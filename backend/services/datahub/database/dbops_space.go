package database

import (
	"log"

	"github.com/blue-monads/turnix/backend/services/datahub/models"
	"github.com/upper/db/v4"
)

func (d *DB) AddSpace(data *models.Space) (int64, error) {
	r, err := d.spaceTable().Insert(data)
	if err != nil {
		return 0, err
	}

	return r.ID().(int64), nil
}

func (d *DB) GetSpace(id int64) (*models.Space, error) {
	data := &models.Space{}

	err := d.spaceTable().Find(db.Cond{"id": id}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *DB) ListSpaces() ([]models.Space, error) {
	datas := make([]models.Space, 0)

	err := d.spaceTable().Find().All(&datas)
	if err != nil {
		return nil, err
	}

	return datas, nil
}

func (d *DB) UpdateSpace(id int64, data map[string]any) error {
	return d.spaceTable().Find(db.Cond{"id": id}).Update(data)
}

func (d *DB) RemoveSpace(id int64) error {
	return d.spaceTable().Find(db.Cond{"id": id}).Delete()
}

func (d *DB) ListSpaceUsers(spaceId int64) ([]models.SpaceUser, error) {

	datas := make([]models.SpaceUser, 0)

	err := d.spaceUserTable().Find(db.Cond{"space_id": spaceId}).All(&datas)
	if err != nil {
		return nil, err
	}

	return datas, nil
}

func (d *DB) AddUserToSpace(ownerId int64, userId int64, spaceId int64) error {

	if !d.isOwner(ownerId, spaceId) {
		return ErrUserNoScope
	}

	_, err := d.spaceUserTable().Insert(models.SpaceUser{
		UserID:  userId,
		SpaceID: spaceId,
	})
	return err
}

func (d *DB) RemoveUserFromSpace(ownerId int64, userId int64, spaceId int64) error {
	return d.spaceUserTable().Find(db.Cond{"owner_id": ownerId, "user_id": userId, "space_id": spaceId}).Delete()
}
func (d *DB) GetSpaceUserScope(userId int64, spaceId int64) (string, error) {

	data := &models.SpaceUser{}
	err := d.spaceUserTable().Find(db.Cond{"user_id": userId, "space_id": spaceId}).One(data)
	if err != nil {
		return "", err
	}

	return data.Scope, nil
}

func (d *DB) ListOwnSpaces(ownerId int64, spaceType string) ([]models.Space, error) {
	cond := db.Cond{"owned_by": ownerId}
	if spaceType != "" {
		cond["stype"] = spaceType
	}

	datas := make([]models.Space, 0)
	err := d.spaceTable().Find(cond).All(&datas)
	if err != nil {
		return nil, err
	}

	return datas, nil
}

type TSpace struct {
	SpaceId int64 `json:"space_id" db:"space_id"`
}

func (d *DB) ListThirdPartySpaces(userId int64, spaceType string) ([]models.Space, error) {
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

	datas := make([]models.Space, 0, len(projs))

	err = d.spaceTable().Find(db.Cond{
		"userId": userId,
		"id IN":  projIds,
	}).All(&datas)

	if err != nil {
		return nil, err
	}

	return datas, nil
}

// Space Configs

func (d *DB) AddSpaceConfig(spaceId int64, uid int64, data *models.SpaceConfig) (int64, error) {
	data.SpaceID = spaceId

	rid, err := d.spaceConfigsTable().Insert(data)
	if err != nil {
		return 0, err
	}
	id := rid.ID().(int64)
	return id, nil
}

func (d *DB) ListSpaceConfigs(spaceId int64) ([]models.SpaceConfig, error) {
	configs := make([]models.SpaceConfig, 0)
	err := d.spaceConfigsTable().Find(db.Cond{"space_id": spaceId}).All(&configs)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func (d *DB) GetSpaceConfig(spaceId int64, uid int64, id int64) (*models.SpaceConfig, error) {
	data := &models.SpaceConfig{}
	err := d.spaceConfigsTable().Find(db.Cond{"space_id": spaceId, "id": id}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *DB) UpdateSpaceConfig(spaceId int64, uid int64, id int64, data map[string]any) error {
	return d.spaceConfigsTable().Find(db.Cond{"space_id": spaceId, "id": id}).Update(&data)
}

func (d *DB) RemoveSpaceConfig(spaceId int64, uid int64, id int64) error {
	return d.spaceConfigsTable().Find(db.Cond{"space_id": spaceId, "id": id}).Delete()
}

// space ops

func (d *DB) ListSpaceTables(spaceId int64) ([]string, error) {
	return nil, nil
}

func (d *DB) ListSpaceTableColumns(spaceId int64, table string) ([]models.SpaceTableColumn, error) {
	return nil, nil
}

func (d *DB) RunSpaceSQLQuery(spaceId int64, query string, data []any) ([]map[string]any, error) {
	return nil, nil
}

func (d *DB) RunSpaceDDL(spaceId int64, ddl string) error {
	return nil
}

// private

func (d *DB) spaceConfigsTable() db.Collection {
	return d.Table("SpaceConfigs")
}

func (d *DB) spaceTable() db.Collection {
	return d.Table("Spaces")
}

func (d *DB) spaceUserTable() db.Collection {
	return d.Table("SpaceUsers")
}

func (d *DB) isOwner(ownerid int64, spaceId int64) bool {
	exist, err := d.spaceTable().Find(db.Cond{"owned_by": ownerid, "id": spaceId}).Exists()

	if err != nil {
		log.Println("owner check error", err)
		return false
	}

	return exist

}
