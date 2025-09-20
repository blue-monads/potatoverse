package actions

import (
	"time"

	"github.com/blue-monads/turnix/backend/services/datahub/models"
	xutils "github.com/blue-monads/turnix/backend/utils"
	"github.com/blue-monads/turnix/backend/utils/libx/easyerr"
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

// User Invites

func (c *Controller) ListUserInvites(offset int, limit int) ([]models.UserInvite, error) {
	return c.database.ListUserInvites(offset, limit)
}

func (c *Controller) GetUserInvite(id int64) (*models.UserInvite, error) {
	return c.database.GetUserInvite(id)
}

func (c *Controller) AddUserInvite(email, role, invitedAsType string, invitedBy int64) (*models.UserInvite, error) {
	// Check if user already exists
	existingUser, err := c.database.GetUserByEmail(email)
	if err == nil && existingUser != nil {
		return nil, easyerr.Error("User with this email already exists")
	}

	// Check if invite already exists
	existingInvite, err := c.database.GetUserInviteByEmail(email)
	if err == nil && existingInvite != nil {
		return nil, easyerr.Error("Invite for this email already exists")
	}

	// Create invite with 7 days expiration
	expiresOn := time.Now().Add(7 * 24 * time.Hour)

	invite := &models.UserInvite{
		Email:         email,
		Role:          role,
		Status:        "pending",
		InvitedBy:     invitedBy,
		InvitedAsType: invitedAsType,
		ExpiresOn:     &expiresOn,
	}

	id, err := c.database.AddUserInvite(invite)
	if err != nil {
		return nil, err
	}

	return c.database.GetUserInvite(id)
}

func (c *Controller) UpdateUserInvite(id int64, data map[string]any) error {
	return c.database.UpdateUserInvite(id, data)
}

func (c *Controller) DeleteUserInvite(id int64) error {
	return c.database.DeleteUserInvite(id)
}

func (c *Controller) ResendUserInvite(id int64) (*models.UserInvite, error) {
	_, err := c.database.GetUserInvite(id)
	if err != nil {
		return nil, err
	}

	// Update expiration to 7 days from now
	expiresOn := time.Now().Add(7 * 24 * time.Hour)

	err = c.database.UpdateUserInvite(id, map[string]any{
		"expires_on": expiresOn,
		"status":     "pending",
	})
	if err != nil {
		return nil, err
	}

	return c.database.GetUserInvite(id)
}

// Create User Directly

func (c *Controller) CreateUserDirectly(name, email, username, utype string, createdBy int64) (*models.User, error) {
	// Check if user already exists by email
	existingUser, err := c.database.GetUserByEmail(email)
	if err == nil && existingUser != nil {
		return nil, easyerr.Error("User with this email already exists")
	}

	// Generate a random password
	password, err := xutils.GenerateRandomString(12)
	if err != nil {
		return nil, err
	}

	// Create new user
	user := &models.User{
		Name:        name,
		Email:       email,
		Username:    &username,
		Utype:       utype,
		Password:    password,
		Bio:         "",
		IsVerified:  false,
		ExtraMeta:   "{}",
		OwnerUserId: createdBy,
		Disabled:    false,
		IsDeleted:   false,
	}

	id, err := c.database.AddUser(user)
	if err != nil {
		return nil, err
	}

	// Get the created user
	createdUser, err := c.database.GetUser(id)
	if err != nil {
		return nil, err
	}

	// Return user with password for display (admin needs to see it)
	return createdUser, nil
}
