package user

import (
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

func (d *UserOperations) AddUserInvite(data *dbmodels.UserInvite) (int64, error) {
	r, err := d.userInviteTable().Insert(data)
	if err != nil {
		return 0, err
	}

	return r.ID().(int64), nil
}

func (d *UserOperations) GetUserInvite(id int64) (*dbmodels.UserInvite, error) {
	data := &dbmodels.UserInvite{}

	err := d.userInviteTable().Find(db.Cond{"id": id}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *UserOperations) GetUserInviteByEmail(email string) (*dbmodels.UserInvite, error) {
	data := &dbmodels.UserInvite{}

	err := d.userInviteTable().Find(db.Cond{"email": email}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *UserOperations) ListUserInvites(offset int, limit int) ([]dbmodels.UserInvite, error) {
	invites := make([]dbmodels.UserInvite, 0)

	err := d.userInviteTable().Find(db.Cond{"id >": offset}).Limit(limit).All(&invites)
	if err != nil {
		return nil, err
	}

	return invites, nil
}

func (d *UserOperations) ListUserInvitesByInviter(inviterId int64) ([]dbmodels.UserInvite, error) {
	invites := make([]dbmodels.UserInvite, 0)

	err := d.userInviteTable().Find(db.Cond{"invited_by": inviterId}).All(&invites)
	if err != nil {
		return nil, err
	}

	return invites, nil
}

func (d *UserOperations) UpdateUserInvite(id int64, data map[string]any) error {
	return d.userInviteTable().Find(db.Cond{"id": id}).Update(data)
}

func (d *UserOperations) DeleteUserInvite(id int64) error {
	return d.userInviteTable().Find(db.Cond{"id": id}).Delete()
}
