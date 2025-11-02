package actions

import (
	"errors"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
)

func (c *Controller) CreateEventSubscription(installId int64, data map[string]any) (*dbmodels.EventSubscription, error) {
	// Validate required fields
	eventKey, ok := data["event_key"].(string)
	if !ok || eventKey == "" {
		return nil, errors.New("event_key is required")
	}

	targetType, ok := data["target_type"].(string)
	if !ok || targetType == "" {
		return nil, errors.New("target_type is required")
	}

	// Extract optional fields
	spaceId, _ := data["space_id"].(float64)
	if spaceId < 0 {
		return nil, errors.New("space_id must be >= 0")
	}

	targetEndpoint, _ := data["target_endpoint"].(string)
	targetOptions, _ := data["target_options"].(string)
	targetCode, _ := data["target_code"].(string)
	rules, _ := data["rules"].(string)
	transform, _ := data["transform"].(string)
	extrameta, _ := data["extrameta"].(string)
	createdBy, _ := data["created_by"].(float64)
	disabled, _ := data["disabled"].(bool)

	eventSub := &dbmodels.EventSubscription{
		SpaceID:        int64(spaceId),
		EventKey:       eventKey,
		TargetType:     targetType,
		TargetEndpoint: targetEndpoint,
		TargetOptions:  targetOptions,
		TargetCode:     targetCode,
		Rules:          rules,
		Transform:      transform,
		ExtraMeta:      extrameta,
		CreatedBy:      int64(createdBy),
		Disabled:       disabled,
	}

	id, err := c.database.GetSpaceOps().AddEventSubscription(installId, eventSub)
	if err != nil {
		return nil, err
	}

	// Get the created event subscription
	return c.database.GetSpaceOps().GetEventSubscription(installId, id)
}

func (c *Controller) UpdateEventSubscriptionByID(installId int64, eventSubscriptionId int64, data map[string]any) (*dbmodels.EventSubscription, error) {
	// Verify the event subscription exists
	_, err := c.GetEventSubscriptionByID(installId, eventSubscriptionId)
	if err != nil {
		return nil, err
	}

	// Update
	err = c.database.GetSpaceOps().UpdateEventSubscription(installId, eventSubscriptionId, data)
	if err != nil {
		return nil, err
	}

	// Return updated entry
	return c.database.GetSpaceOps().GetEventSubscription(installId, eventSubscriptionId)
}

func (c *Controller) DeleteEventSubscriptionByID(installId int64, eventSubscriptionId int64) error {
	return c.database.GetSpaceOps().RemoveEventSubscription(installId, eventSubscriptionId)
}

func (c *Controller) QueryEventSubscriptions(installId int64, cond map[any]any) ([]dbmodels.EventSubscription, error) {
	return c.database.GetSpaceOps().QueryEventSubscriptions(installId, cond)
}

func (c *Controller) GetEventSubscriptionByID(installId int64, eventSubscriptionId int64) (*dbmodels.EventSubscription, error) {
	return c.database.GetSpaceOps().GetEventSubscription(installId, eventSubscriptionId)
}


