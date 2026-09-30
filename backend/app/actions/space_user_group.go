package actions

import (
	"errors"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
)

func (c *Controller) CreateSpaceUserGroup(installId int64, data map[string]any) (*dbmodels.SpaceUserGroup, error) {
	var groupId int64
	switch v := data["group_id"].(type) {
	case float64:
		groupId = int64(v)
	case int64:
		groupId = v
	case int:
		groupId = int64(v)
	}

	if groupId <= 0 {
		return nil, errors.New("group_id is required")
	}

	var spaceId int64
	switch v := data["space_id"].(type) {
	case float64:
		spaceId = int64(v)
	case int64:
		spaceId = v
	case int:
		spaceId = int64(v)
	}
	if spaceId < 0 {
		return nil, errors.New("space_id must be >= 0")
	}

	scope, _ := data["scope"].(string)
	extrameta, _ := data["extrameta"].(string)

	spaceUserGroup := &dbmodels.SpaceUserGroup{
		GroupID:   groupId,
		SpaceID:   spaceId,
		Scope:     scope,
		ExtraMeta: extrameta,
	}

	id, err := c.database.GetSpaceOps().AddSpaceUserGroup(installId, spaceUserGroup)
	if err != nil {
		return nil, err
	}

	return c.database.GetSpaceOps().GetSpaceUserGroup(installId, id)
}

func (c *Controller) UpdateSpaceUserGroupByID(installId int64, spaceUserGroupId int64, data map[string]any) (*dbmodels.SpaceUserGroup, error) {
	_, err := c.GetSpaceUserGroupByID(installId, spaceUserGroupId)
	if err != nil {
		return nil, err
	}

	err = c.database.GetSpaceOps().UpdateSpaceUserGroup(installId, spaceUserGroupId, data)
	if err != nil {
		return nil, err
	}

	return c.database.GetSpaceOps().GetSpaceUserGroup(installId, spaceUserGroupId)
}

func (c *Controller) DeleteSpaceUserGroupByID(installId int64, spaceUserGroupId int64) error {
	return c.database.GetSpaceOps().RemoveSpaceUserGroup(installId, spaceUserGroupId)
}

func (c *Controller) QuerySpaceUserGroups(installId int64, cond map[any]any) ([]dbmodels.SpaceUserGroup, error) {
	return c.database.GetSpaceOps().QuerySpaceUserGroups(installId, cond)
}

func (c *Controller) GetSpaceUserGroupByID(installId int64, spaceUserGroupId int64) (*dbmodels.SpaceUserGroup, error) {
	return c.database.GetSpaceOps().GetSpaceUserGroup(installId, spaceUserGroupId)
}
