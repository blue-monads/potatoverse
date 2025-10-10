package actions

import (
	"github.com/blue-monads/turnix/backend/services/datahub/models"
)

// User Group actions

func (c *Controller) ListUserGroups() ([]models.UserGroup, error) {
	return c.database.ListUserGroups()
}

func (c *Controller) GetUserGroup(name string) (*models.UserGroup, error) {
	return c.database.GetUserGroup(name)
}

func (c *Controller) AddUserGroup(name string, info string) error {
	return c.database.AddUserGroup(name, info)
}

func (c *Controller) UpdateUserGroup(name string, info string) error {
	return c.database.UpdateUserGroup(name, info)
}

func (c *Controller) DeleteUserGroup(name string) error {
	return c.database.DeleteUserGroup(name)
}
