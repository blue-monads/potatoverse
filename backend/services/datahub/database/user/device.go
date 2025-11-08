package user

import (
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

func (d *UserOperations) ListUserDevice(userId int64) ([]dbmodels.UserDevice, error) {

	devices := make([]dbmodels.UserDevice, 0)

	err := d.deviceTable().Find(db.Cond{"user_id": userId}).All(&devices)
	if err != nil {
		return nil, err
	}

	return devices, nil
}

func (d *UserOperations) AddUserDevice(data *dbmodels.UserDevice) (int64, error) {
	r, err := d.deviceTable().Insert(data)
	if err != nil {
		return 0, err
	}

	return r.ID().(int64), nil
}

func (d *UserOperations) GetUserDevice(id int64) (*dbmodels.UserDevice, error) {

	data := &dbmodels.UserDevice{}

	err := d.deviceTable().Find(db.Cond{"id": id}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *UserOperations) UpdateUserDevice(id int64, data map[string]any) error {
	return d.deviceTable().Find(db.Cond{"id": id}).Update(data)
}

func (d *UserOperations) DeleteUserDevice(id int64) error {
	return d.deviceTable().Find(db.Cond{"id": id}).Delete()
}

// private

func (d *UserOperations) deviceTable() db.Collection {
	return d.db.Collection("UserDevices")
}

func (d *UserOperations) userTable() db.Collection {
	return d.db.Collection("Users")
}

func (d *UserOperations) userInviteTable() db.Collection {
	return d.db.Collection("UserInvites")
}

func (d *UserOperations) userGroupTable() db.Collection {
	return d.db.Collection("UserGroups")
}
