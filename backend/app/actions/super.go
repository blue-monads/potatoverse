package actions

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	xutils "github.com/blue-monads/potatoverse/backend/utils"
)

func (c *Controller) GetAdminToken() (*dbmodels.User, string, error) {
	users, err := c.ListUsers(0, 200)
	if err != nil {
		return nil, "", err
	}

	var admin *dbmodels.User
	for i := range users {
		u := &users[i]
		if u.Ugroup == "admin" && !u.Disabled && !u.IsDeleted {
			admin = u
			break
		}
	}
	if admin == nil {
		return nil, "", errors.New("no admin user found")
	}

	token, err := c.signer.SignAccess(&signer.AccessClaim{
		UserId: admin.ID,
	})
	if err != nil {
		return nil, "", err
	}

	admin.Password = ""
	admin.ExtraMeta = ""

	return admin, token, nil
}

func (c *Controller) ListAdminUsers() ([]dbmodels.User, error) {
	users, err := c.ListUsers(0, 500)
	if err != nil {
		return nil, err
	}

	admins := make([]dbmodels.User, 0)
	for i := range users {
		u := users[i]
		if u.Ugroup == "admin" && !u.IsDeleted {
			u.Password = ""
			u.ExtraMeta = ""
			admins = append(admins, u)
		}
	}
	return admins, nil
}

func (c *Controller) ListAdminUser() ([]dbmodels.User, error) {
	return c.ListAdminUsers()
}

func (c *Controller) ListAllUsers() ([]dbmodels.User, error) {
	users, err := c.ListUsers(0, 500)
	if err != nil {
		return nil, err
	}

	all := make([]dbmodels.User, 0, len(users))
	for i := range users {
		u := users[i]
		if !u.IsDeleted {
			u.Password = ""
			u.ExtraMeta = ""
			all = append(all, u)
		}
	}
	return all, nil
}

func (c *Controller) ListAllUser() ([]dbmodels.User, error) {
	return c.ListAllUsers()
}

func (c *Controller) ResetUserPass(userOrId string, newPassword string) (*dbmodels.User, string, error) {
	var targetUser *dbmodels.User

	userOrId = strings.TrimSpace(userOrId)
	if userOrId == "" {
		// Default to primary active admin user
		users, err := c.ListUsers(0, 200)
		if err != nil {
			return nil, "", err
		}
		for i := range users {
			u := &users[i]
			if u.Ugroup == "admin" && !u.Disabled && !u.IsDeleted {
				targetUser = u
				break
			}
		}
		if targetUser == nil {
			return nil, "", errors.New("no active admin user found")
		}
	} else if id, err := strconv.ParseInt(userOrId, 10, 64); err == nil && id > 0 {
		u, err := c.database.GetUserOps().GetUser(id)
		if err != nil {
			return nil, "", err
		}
		targetUser = u
	} else {
		// Try username
		u, err := c.database.GetUserOps().GetUserByUsername(userOrId)
		if err == nil && u != nil && u.ID > 0 {
			targetUser = u
		} else {
			// Try email
			u, err = c.database.GetUserOps().GetUserByEmail(userOrId)
			if err != nil {
				return nil, "", fmt.Errorf("user not found: %s", userOrId)
			}
			targetUser = u
		}
	}

	if targetUser == nil {
		return nil, "", errors.New("user not found")
	}
	if targetUser.IsDeleted {
		return nil, "", errors.New("user is deleted")
	}

	password := strings.TrimSpace(newPassword)
	if password == "" {
		randPass, err := xutils.GenerateRandomString(12)
		if err != nil {
			return nil, "", err
		}
		password = randPass
	}

	hashedPassword, err := xutils.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	err = c.database.GetUserOps().UpdateUser(targetUser.ID, map[string]any{
		"password": hashedPassword,
	})
	if err != nil {
		return nil, "", err
	}

	targetUser.Password = ""
	targetUser.ExtraMeta = ""

	return targetUser, password, nil
}