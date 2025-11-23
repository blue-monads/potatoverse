package actions

import (
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
)

func (c *Controller) CreateEventSubscription(installId int64, data *dbmodels.MQSubscription) (*dbmodels.MQSubscription, error) {

	id, err := c.database.GetSpaceOps().AddEventSubscription(installId, data)
	if err != nil {
		return nil, err
	}

	// Get the created event subscription
	c.engine.RefreshEventIndex()

	return c.database.GetSpaceOps().GetEventSubscription(installId, id)
}

func (c *Controller) UpdateEventSubscriptionByID(installId int64, eventSubscriptionId int64, data map[string]any) (*dbmodels.MQSubscription, error) {
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
	err := c.database.GetSpaceOps().RemoveEventSubscription(installId, eventSubscriptionId)
	if err != nil {
		return err
	}

	c.engine.RefreshEventIndex()

	return nil
}

func (c *Controller) QueryEventSubscriptions(installId int64, cond map[any]any) ([]dbmodels.MQSubscription, error) {
	return c.database.GetSpaceOps().QueryEventSubscriptions(installId, cond)
}

func (c *Controller) GetEventSubscriptionByID(installId int64, eventSubscriptionId int64) (*dbmodels.MQSubscription, error) {
	return c.database.GetSpaceOps().GetEventSubscription(installId, eventSubscriptionId)
}
