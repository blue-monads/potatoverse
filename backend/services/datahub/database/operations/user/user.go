package user

import (
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

type UserOperations struct {
	db db.Session
}

func NewUserOperations(db db.Session) *UserOperations {
	return &UserOperations{
		db: db,
	}
}

func (d *UserOperations) AddUser(data *dbmodels.User) (int64, error) {
	r, err := d.userTable().Insert(data)
	if err != nil {
		return 0, err
	}

	return r.ID().(int64), nil
}

func (d *UserOperations) UpdateUser(id int64, data map[string]any) error {
	return d.userTable().Find(db.Cond{"id": id}).Update(data)
}

func (d *UserOperations) GetUser(id int64) (*dbmodels.User, error) {

	data := &dbmodels.User{}

	err := d.userTable().Find(db.Cond{"id": id}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *UserOperations) GetUserByEmail(email string) (*dbmodels.User, error) {

	data := &dbmodels.User{}

	err := d.userTable().Find(db.Cond{"email": email}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *UserOperations) GetUserByUsername(username string) (*dbmodels.User, error) {

	data := &dbmodels.User{}

	err := d.userTable().Find(db.Cond{"username": username}).One(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (d *UserOperations) ListUser(offset int, limit int) ([]dbmodels.User, error) {

	users := make([]dbmodels.User, 0)

	err := d.userTable().Find(db.Cond{"id >": offset}).Limit(limit).All(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (d *UserOperations) ListUserByOwner(owner int64) ([]dbmodels.User, error) {

	users := make([]dbmodels.User, 0)

	err := d.userTable().Find(db.Cond{"owner_user_id": owner}).All(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (d *UserOperations) DeleteUser(id int64) error {
	return d.userTable().Find(db.Cond{"id": id}).Delete()
}
