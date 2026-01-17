package actions

import (
	"errors"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
)

func (c *Controller) CreateSpaceKV(installId int64, data map[string]any) (*dbmodels.SpaceKV, error) {
	// Validate required fields
	key, ok := data["key"].(string)
	if !ok || key == "" {
		return nil, errors.New("key is required")
	}

	groupName, ok := data["group"].(string)
	if !ok || groupName == "" {
		return nil, errors.New("group is required")
	}

	value, ok := data["value"].(string)
	if !ok || value == "" {
		return nil, errors.New("value is required")
	}

	// Extract optional fields
	tag1, _ := data["tag1"].(string)
	tag2, _ := data["tag2"].(string)
	tag3, _ := data["tag3"].(string)

	kv := &dbmodels.SpaceKV{
		InstallID: installId,
		Key:       key,
		Group:     groupName,
		Value:     value,
		Tag1:      tag1,
		Tag2:      tag2,
		Tag3:      tag3,
	}

	err := c.database.GetSpaceKVOps().AddSpaceKV(installId, kv)
	if err != nil {
		return nil, err
	}

	// Get the created KV entry
	return c.database.GetSpaceKVOps().GetSpaceKV(installId, groupName, key)
}

func (c *Controller) UpdateSpaceKVByID(installId int64, kvId int64, data map[string]any) (*dbmodels.SpaceKV, error) {
	// Get the existing KV entry
	kv, err := c.GetSpaceKVByID(installId, kvId)
	if err != nil {
		return nil, err
	}

	// Update using group and key
	err = c.database.GetSpaceKVOps().UpdateSpaceKV(installId, kv.Group, kv.Key, data)
	if err != nil {
		return nil, err
	}

	// Return updated entry
	return c.database.GetSpaceKVOps().GetSpaceKV(installId, kv.Group, kv.Key)
}

func (c *Controller) DeleteSpaceKVByID(installId int64, kvId int64) error {
	// Get the existing KV entry to find group and key
	kv, err := c.GetSpaceKVByID(installId, kvId)
	if err != nil {
		return err
	}

	return c.database.GetSpaceKVOps().RemoveSpaceKV(installId, kv.Group, kv.Key)
}

func (c *Controller) QuerySpaceKV(installId int64, cond map[any]any, offset int, limit int) ([]dbmodels.SpaceKV, error) {
	return c.database.GetSpaceKVOps().QuerySpaceKV(installId, cond, offset, limit)
}

func (c *Controller) GetSpaceKVByID(installId int64, kvId int64) (*dbmodels.SpaceKV, error) {
	return c.database.GetSpaceKVOps().GetSpaceKVByID(installId, kvId)
}
