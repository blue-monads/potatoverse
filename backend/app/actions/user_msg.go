package actions

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
)

// UserMessage actions

func (c *Controller) GetUserMessage(id int64) (*dbmodels.UserMessage, error) {
	return c.database.GetUserOps().GetUserMessage(id)
}

func (c *Controller) ListUserMessages(toUserId int64, afterId int64, limit int) ([]dbmodels.UserMessage, error) {
	return c.database.GetUserOps().ListUserMessages(toUserId, afterId, limit)
}

func (c *Controller) QueryNewMessages(toUserId int64, readHead int64) ([]dbmodels.UserMessage, error) {
	return c.database.GetUserOps().QueryNewMessages(toUserId, readHead)
}

func (c *Controller) GetUserReadHead(userId int64) (int64, error) {
	user, err := c.database.GetUserOps().GetUser(userId)
	if err != nil {
		return 0, err
	}
	return user.MessageReadHead, nil
}

func (c *Controller) QueryMessageHistory(toUserId int64, limit int) ([]dbmodels.UserMessage, error) {
	return c.database.GetUserOps().QueryMessageHistory(toUserId, limit)
}

func (c *Controller) SetMessageAsRead(id int64, toUserId int64) error {
	return c.database.GetUserOps().SetMessageAsRead(id, toUserId)
}

func (c *Controller) SetAllMessagesAsRead(toUserId int64, readHead int64) error {
	return c.database.GetUserOps().SetAllMessagesAsRead(toUserId, readHead)
}

func (c *Controller) UpdateUserMessage(id int64, data map[string]any) error {
	return c.database.GetUserOps().UpdateUserMessage(id, data)
}

func (c *Controller) DeleteUserMessage(id int64) error {
	return c.database.GetUserOps().DeleteUserMessage(id)
}
