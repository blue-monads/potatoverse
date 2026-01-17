package user

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

// user groups

func (d *UserOperations) AddUserGroup(name string, info string) error {
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

func (d *UserOperations) GetUserGroup(name string) (*dbmodels.UserGroup, error) {
	data := &dbmodels.UserGroup{}

	err := d.userGroupTable().Find(db.Cond{"name": name}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *UserOperations) ListUserGroups() ([]dbmodels.UserGroup, error) {
	userGroups := make([]dbmodels.UserGroup, 0)

	err := d.userGroupTable().Find().All(&userGroups)
	if err != nil {
		return nil, err
	}

	return userGroups, nil
}

func (d *UserOperations) UpdateUserGroup(name string, info string) error {
	return d.userGroupTable().
		Find(db.Cond{"name": name}).
		Update(map[string]any{
			"info": info,
		})
}

func (d *UserOperations) DeleteUserGroup(name string) error {
	return d.userGroupTable().Find(db.Cond{"name": name}).Delete()
}
