package actions

import (
	"errors"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
)

func (c *Controller) CreateSpaceUser(installId int64, data map[string]any) (*dbmodels.SpaceUser, error) {
	// Validate required fields
	userId, ok := data["user_id"].(float64)
	if !ok || userId == 0 {
		return nil, errors.New("user_id is required")
	}

	spaceId, _ := data["space_id"].(float64)
	if spaceId < 0 {
		return nil, errors.New("space_id must be >= 0")
	}

	// Extract optional fields
	scope, _ := data["scope"].(string)
	token, _ := data["token"].(string)
	extrameta, _ := data["extrameta"].(string)

	spaceUser := &dbmodels.SpaceUser{
		UserID:    int64(userId),
		SpaceID:   int64(spaceId),
		Scope:     scope,
		Token:     token,
		ExtraMeta: extrameta,
	}

	id, err := c.database.GetSpaceOps().AddSpaceUser(installId, spaceUser)
	if err != nil {
		return nil, err
	}

	// Get the created space user
	return c.database.GetSpaceOps().GetSpaceUser(installId, id)
}

func (c *Controller) UpdateSpaceUserByID(installId int64, spaceUserId int64, data map[string]any) (*dbmodels.SpaceUser, error) {
	// Verify the space user exists
	_, err := c.GetSpaceUserByID(installId, spaceUserId)
	if err != nil {
		return nil, err
	}

	// Update
	err = c.database.GetSpaceOps().UpdateSpaceUser(installId, spaceUserId, data)
	if err != nil {
		return nil, err
	}

	// Return updated entry
	return c.database.GetSpaceOps().GetSpaceUser(installId, spaceUserId)
}

func (c *Controller) DeleteSpaceUserByID(installId int64, spaceUserId int64) error {
	return c.database.GetSpaceOps().RemoveSpaceUser(installId, spaceUserId)
}

func (c *Controller) QuerySpaceUsers(installId int64, cond map[any]any) ([]dbmodels.SpaceUser, error) {
	return c.database.GetSpaceOps().QuerySpaceUsers(installId, cond)
}

func (c *Controller) GetSpaceUserByID(installId int64, spaceUserId int64) (*dbmodels.SpaceUser, error) {
	return c.database.GetSpaceOps().GetSpaceUser(installId, spaceUserId)
}

