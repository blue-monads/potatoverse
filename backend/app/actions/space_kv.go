package actions

import (
	"errors"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
)

func (c *Controller) CreateSpaceKV(spaceId int64, data map[string]any) (*dbmodels.SpaceKV, error) {
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
		SpaceID: spaceId,
		Key:     key,
		Group:   groupName,
		Value:   value,
		Tag1:    tag1,
		Tag2:    tag2,
		Tag3:    tag3,
	}

	err := c.database.AddSpaceKV(spaceId, kv)
	if err != nil {
		return nil, err
	}

	// Get the created KV entry
	return c.database.GetSpaceKV(spaceId, groupName, key)
}

func (c *Controller) UpdateSpaceKVByID(spaceId, kvId int64, data map[string]any) (*dbmodels.SpaceKV, error) {
	// Get the existing KV entry
	kv, err := c.GetSpaceKVByID(spaceId, kvId)
	if err != nil {
		return nil, err
	}

	// Update using group and key
	err = c.database.UpdateSpaceKV(spaceId, kv.Group, kv.Key, data)
	if err != nil {
		return nil, err
	}

	// Return updated entry
	return c.database.GetSpaceKV(spaceId, kv.Group, kv.Key)
}

func (c *Controller) DeleteSpaceKVByID(spaceId, kvId int64) error {
	// Get the existing KV entry to find group and key
	kv, err := c.GetSpaceKVByID(spaceId, kvId)
	if err != nil {
		return err
	}

	return c.database.RemoveSpaceKV(spaceId, kv.Group, kv.Key)
}

func (c *Controller) GetSpace(spaceId int64) (*dbmodels.Space, error) {
	return c.database.GetSpace(spaceId)
}

func (c *Controller) QuerySpaceKV(spaceId int64, cond map[any]any) ([]dbmodels.SpaceKV, error) {
	return c.database.QuerySpaceKV(spaceId, cond)
}

func (c *Controller) GetSpaceKVByID(spaceId, kvId int64) (*dbmodels.SpaceKV, error) {
	// First get all KV entries for the space and find by ID
	kvEntries, err := c.database.QuerySpaceKV(spaceId, map[any]any{})
	if err != nil {
		return nil, err
	}

	for _, kv := range kvEntries {
		if kv.ID == kvId {
			return &kv, nil
		}
	}

	return nil, errors.New("KV entry not found")
}
