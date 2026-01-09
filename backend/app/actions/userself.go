package actions

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
)

// GetSelfInfo returns the current user's information
func (c *Controller) GetSelfInfo(userId int64) (*dbmodels.User, error) {
	user, err := c.database.GetUserOps().GetUser(userId)
	if err != nil {
		return nil, err
	}

	// Remove sensitive information
	user.Password = ""
	user.ExtraMeta = ""
	user.OwnerUserId = 0
	user.OwnerSpaceId = 0
	user.MessageReadHead = 0
	user.Disabled = false
	user.IsDeleted = false

	return user, nil
}

// UpdateSelfBio updates the current user's bio
func (c *Controller) UpdateSelfBio(userId int64, bio string) error {
	return c.database.GetUserOps().UpdateUser(userId, map[string]any{
		"bio": bio,
	})
}

/*

todo:
- update email
- update password
- message another user
- read self messages
- delete self messages
- update read status of self messages

*/
