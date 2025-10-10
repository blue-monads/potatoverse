package actions

import (
	"time"

	"github.com/blue-monads/turnix/backend/services/datahub/models"
	"github.com/blue-monads/turnix/backend/services/mailer"
	"github.com/blue-monads/turnix/backend/services/signer"
	xutils "github.com/blue-monads/turnix/backend/utils"
	"github.com/blue-monads/turnix/backend/utils/libx/easyerr"
	"github.com/k0kubun/pp"
	"github.com/rs/xid"
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

type UserInviteResponse struct {
	*models.UserInvite
	InviteUrl string `json:"invite_url"`
}

func (c *Controller) AddUserInvite(email, role, invitedAsType string, invitedBy int64) (*UserInviteResponse, error) {
	existingUser, err := c.database.GetUserByEmail(email)
	if err == nil && existingUser != nil {
		return nil, easyerr.Error("User with this email already exists")
	}

	existingInvite, err := c.database.GetUserInviteByEmail(email)
	if err == nil && existingInvite != nil {
		return nil, easyerr.Error("Invite for this email already exists")
	}

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

	inviteToken, err := c.signer.SignInvite(&signer.InviteClaim{
		XID:      xid.New().String(),
		Typeid:   signer.TokenTypeEmailInvite,
		InviteId: id,
	})
	if err != nil {
		return nil, err
	}

	host := c.AppOpts.Host
	port := c.AppOpts.Port

	pp.Println(host, port)

	fullUrl := xutils.GetFullUrl(host, "/zz/pages/auth/signup/invite-finish?token="+inviteToken, port, false)

	body := &mailer.SimpleMessage{
		Text: fullUrl,
		HTML: `<h1>Welcome to Turnix</h1><p> Please click the link below to accept the invite: <a href="` + fullUrl + `">` + fullUrl + `</a></p>`,
	}

	err = c.mailer.Send(email, "Welcome to Turnix", body)
	if err != nil {
		return nil, err
	}

	inviteData, err := c.database.GetUserInvite(id)
	if err != nil {
		return nil, err
	}

	return &UserInviteResponse{
		UserInvite: inviteData,
		InviteUrl:  fullUrl,
	}, nil
}

func (c *Controller) UpdateUserInvite(id int64, data map[string]any) error {
	return c.database.UpdateUserInvite(id, data)
}

func (c *Controller) DeleteUserInvite(id int64) error {
	return c.database.DeleteUserInvite(id)
}

func (c *Controller) ResendUserInvite(id int64) (*UserInviteResponse, error) {
	invite, err := c.database.GetUserInvite(id)
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

	// Generate new invite token
	inviteToken, err := c.signer.SignInvite(&signer.InviteClaim{
		XID:      xid.New().String(),
		Typeid:   signer.TokenTypeEmailInvite,
		InviteId: id,
	})
	if err != nil {
		return nil, err
	}

	host := c.AppOpts.Host
	port := c.AppOpts.Port
	fullUrl := xutils.GetFullUrl(host, "/zz/pages/auth/signup/invite-finish?token="+inviteToken, port, false)

	// Send email
	body := &mailer.SimpleMessage{
		Text: fullUrl,
		HTML: `<h1>Welcome to Turnix</h1><p> Please click the link below to accept the invite: <a href="` + fullUrl + `">` + fullUrl + `</a></p>`,
	}

	err = c.mailer.Send(invite.Email, "Welcome to Turnix", body)
	if err != nil {
		return nil, err
	}

	updatedInvite, err := c.database.GetUserInvite(id)
	if err != nil {
		return nil, err
	}

	return &UserInviteResponse{
		UserInvite: updatedInvite,
		InviteUrl:  fullUrl,
	}, nil
}

func (c *Controller) AcceptUserInvite(inviteId int64, name, username, password string) (*models.User, error) {
	// Get the invite
	invite, err := c.database.GetUserInvite(inviteId)
	if err != nil {
		return nil, easyerr.Error("Invalid invite")
	}

	// Check if invite is still pending
	if invite.Status != "pending" {
		return nil, easyerr.Error("Invite has already been used or expired")
	}

	// Check if invite has expired
	if invite.ExpiresOn != nil && time.Now().After(*invite.ExpiresOn) {
		return nil, easyerr.Error("Invite has expired")
	}

	// Check if user already exists by email
	existingUser, err := c.database.GetUserByEmail(invite.Email)
	if err == nil && existingUser != nil {
		return nil, easyerr.Error("User with this email already exists")
	}

	// Check if username is already taken
	existingUserByUsername, err := c.database.GetUserByUsername(username)
	if err == nil && existingUserByUsername != nil {
		return nil, easyerr.Error("Username is already taken")
	}

	// Create new user
	user := &models.User{
		Name:        name,
		Email:       invite.Email,
		Username:    &username,
		Utype:       "user",
		Ugroup:      invite.InvitedAsType,
		Password:    password,
		Bio:         "",
		IsVerified:  true, // Invited users are considered verified
		ExtraMeta:   "{}",
		OwnerUserId: invite.InvitedBy,
	}

	userId, err := c.database.AddUser(user)
	if err != nil {
		return nil, err
	}

	// Update invite status to accepted
	err = c.database.UpdateUserInvite(inviteId, map[string]any{
		"status": "accepted",
	})
	if err != nil {
		return nil, err
	}

	return c.database.GetUser(userId)
}

// Create User Directly

func (c *Controller) CreateUserDirectly(name, email, username, utype, ugroup string, createdBy int64) (*models.User, error) {
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
		Ugroup:      ugroup,
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
