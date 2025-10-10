package database

import (
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

var _ datahub.UserOps = (*DB)(nil)

func (d *DB) AddUser(data *dbmodels.User) (int64, error) {
	r, err := d.userTable().Insert(data)
	if err != nil {
		return 0, err
	}

	return r.ID().(int64), nil
}

func (d *DB) UpdateUser(id int64, data map[string]any) error {
	return d.userTable().Find(db.Cond{"id": id}).Update(data)
}

func (d *DB) GetUser(id int64) (*dbmodels.User, error) {

	data := &dbmodels.User{}

	err := d.userTable().Find(db.Cond{"id": id}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *DB) GetUserByEmail(email string) (*dbmodels.User, error) {

	data := &dbmodels.User{}

	err := d.userTable().Find(db.Cond{"email": email}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *DB) GetUserByUsername(username string) (*dbmodels.User, error) {

	data := &dbmodels.User{}

	err := d.userTable().Find(db.Cond{"username": username}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *DB) ListUser(offset int, limit int) ([]dbmodels.User, error) {

	users := make([]dbmodels.User, 0)

	err := d.userTable().Find(db.Cond{"id >": offset}).Limit(limit).All(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (d *DB) ListUserByOwner(owner int64) ([]dbmodels.User, error) {

	users := make([]dbmodels.User, 0)

	err := d.userTable().Find(db.Cond{"owner_user_id": owner}).All(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (d *DB) DeleteUser(id int64) error {
	return d.userTable().Find(db.Cond{"id": id}).Delete()
}

// devices

func (d *DB) ListUserDevice(userId int64) ([]dbmodels.UserDevice, error) {

	devices := make([]dbmodels.UserDevice, 0)

	err := d.deviceTable().Find(db.Cond{"user_id": userId}).All(&devices)
	if err != nil {
		return nil, err
	}

	return devices, nil
}

func (d *DB) AddUserDevice(data *dbmodels.UserDevice) (int64, error) {
	r, err := d.deviceTable().Insert(data)
	if err != nil {
		return 0, err
	}

	return r.ID().(int64), nil
}

func (d *DB) GetUserDevice(id int64) (*dbmodels.UserDevice, error) {

	data := &dbmodels.UserDevice{}

	err := d.deviceTable().Find(db.Cond{"id": id}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *DB) UpdateUserDevice(id int64, data map[string]any) error {
	return d.deviceTable().Find(db.Cond{"id": id}).Update(data)
}

func (d *DB) DeleteUserDevice(id int64) error {
	return d.deviceTable().Find(db.Cond{"id": id}).Delete()
}

// user invites

func (d *DB) AddUserInvite(data *dbmodels.UserInvite) (int64, error) {
	r, err := d.userInviteTable().Insert(data)
	if err != nil {
		return 0, err
	}

	return r.ID().(int64), nil
}

func (d *DB) GetUserInvite(id int64) (*dbmodels.UserInvite, error) {
	data := &dbmodels.UserInvite{}

	err := d.userInviteTable().Find(db.Cond{"id": id}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *DB) GetUserInviteByEmail(email string) (*dbmodels.UserInvite, error) {
	data := &dbmodels.UserInvite{}

	err := d.userInviteTable().Find(db.Cond{"email": email}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *DB) ListUserInvites(offset int, limit int) ([]dbmodels.UserInvite, error) {
	invites := make([]dbmodels.UserInvite, 0)

	err := d.userInviteTable().Find(db.Cond{"id >": offset}).Limit(limit).All(&invites)
	if err != nil {
		return nil, err
	}

	return invites, nil
}

func (d *DB) ListUserInvitesByInviter(inviterId int64) ([]dbmodels.UserInvite, error) {
	invites := make([]dbmodels.UserInvite, 0)

	err := d.userInviteTable().Find(db.Cond{"invited_by": inviterId}).All(&invites)
	if err != nil {
		return nil, err
	}

	return invites, nil
}

func (d *DB) UpdateUserInvite(id int64, data map[string]any) error {
	return d.userInviteTable().Find(db.Cond{"id": id}).Update(data)
}

func (d *DB) DeleteUserInvite(id int64) error {
	return d.userInviteTable().Find(db.Cond{"id": id}).Delete()
}

// user groups

func (d *DB) AddUserGroup(name string, info string) error {
	userGroup := &dbmodels.UserGroup{
		Name: name,
		Info: info,
	}

	_, err := d.userGroupTable().Insert(userGroup)
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) GetUserGroup(name string) (*dbmodels.UserGroup, error) {
	data := &dbmodels.UserGroup{}

	err := d.userGroupTable().Find(db.Cond{"name": name}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *DB) ListUserGroups() ([]dbmodels.UserGroup, error) {
	userGroups := make([]dbmodels.UserGroup, 0)

	err := d.userGroupTable().Find().All(&userGroups)
	if err != nil {
		return nil, err
	}

	return userGroups, nil
}

func (d *DB) UpdateUserGroup(name string, info string) error {
	return d.userGroupTable().Find(db.Cond{"name": name}).Update(map[string]any{
		"info": info,
	})
}

func (d *DB) DeleteUserGroup(name string) error {
	return d.userGroupTable().Find(db.Cond{"name": name}).Delete()
}

// private

func (d *DB) deviceTable() db.Collection {
	return d.Table("UserDevices")
}

func (d *DB) userTable() db.Collection {
	return d.Table("Users")
}

func (d *DB) userInviteTable() db.Collection {
	return d.Table("UserInvites")
}

func (d *DB) userGroupTable() db.Collection {
	return d.Table("UserGroups")
}
