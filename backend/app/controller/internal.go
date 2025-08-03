package controller

import (
	"github.com/blue-monads/turnix/backend/services/datahub/models"
)

func (c *Controller) AddAdminUserDirect(name, password string) (*models.User, error) {
	return c.AddUserDirect(name, password, "admin")
}

func (c *Controller) AddNormalUserDirect(name, password string) (*models.User, error) {
	return c.AddUserDirect(name, password, "normal")
}

func (c *Controller) AddUserDirect(name, password, utype string) (*models.User, error) {

	uid, err := c.database.AddUser(&models.User{
		ID:         0,
		Name:       name,
		Bio:        "This is a normal user.",
		Utype:      utype,
		Username:   &name,
		IsVerified: true,
		Password:   password,
	})
	if err != nil {
		return nil, err
	}

	return c.database.GetUser(uid)
}
