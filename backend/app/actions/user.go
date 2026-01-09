package actions

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
)

// User Group actions

func (c *Controller) ListUserGroups() ([]dbmodels.UserGroup, error) {
	return c.database.GetUserOps().ListUserGroups()
}

func (c *Controller) GetUserGroup(name string) (*dbmodels.UserGroup, error) {
	return c.database.GetUserOps().GetUserGroup(name)
}

func (c *Controller) AddUserGroup(name string, info string) error {
	return c.database.GetUserOps().AddUserGroup(name, info)
}

func (c *Controller) UpdateUserGroup(name string, info string) error {
	return c.database.GetUserOps().UpdateUserGroup(name, info)
}

func (c *Controller) DeleteUserGroup(name string) error {
	return c.database.GetUserOps().DeleteUserGroup(name)
}
