package actions

import (
	"github.com/blue-monads/turnix/backend/services/datahub/models"
	xutils "github.com/blue-monads/turnix/backend/utils"
)

func (c *Controller) ListUsers(offset int, limit int) ([]models.User, error) {
	return c.database.ListUser(offset, limit)
}

func (c *Controller) AddUser(user *models.User) (int64, error) {
	return c.database.AddUser(user)
}

func (c *Controller) GetUser(id int64) (*models.User, error) {
	usr, err := c.database.GetUser(id)
	if err != nil {
		return nil, err
	}

	usr.Password = ""
	usr.ExtraMeta = ""
	usr.CreatedAt = nil
	usr.OwnerUserId = 0
	usr.OwnerSpaceId = 0
	usr.MessageReadHead = 0
	usr.Disabled = false
	usr.IsDeleted = false

	return usr, nil
}

func (c *Controller) ResetUserPassword(id int64) (string, error) {

	user, err := c.database.GetUser(id)
	if err != nil {
		return "", err
	}

	password, err := xutils.GenerateRandomString(10)
	if err != nil {
		return "", err
	}

	user.Password = password

	err = c.database.UpdateUser(id, map[string]any{
		"password": password,
	})
	if err != nil {
		return "", err
	}

	return password, nil
}

func (c *Controller) DeactivateUser(id int64) error {
	return c.database.UpdateUser(id, map[string]any{
		"disabled": true,
	})
}

func (c *Controller) ActivateUser(id int64) error {
	return c.database.UpdateUser(id, map[string]any{
		"disabled": false,
	})
}

func (c *Controller) DeleteUser(id int64) error {
	return c.database.DeleteUser(id)
}

func (c *Controller) UpdateUser(id int64, user *models.User) error {

	return c.database.UpdateUser(id, map[string]any{
		"name": user.Name,
		"bio":  user.Bio,
	})
}
