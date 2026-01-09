package user

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

// UserMessage operations

func (d *UserOperations) AddUserMessage(data *dbmodels.UserMessage) (int64, error) {
	r, err := d.userMessageTable().Insert(data)
	if err != nil {
		return 0, err
	}
	return r.ID().(int64), nil
}

func (d *UserOperations) GetUserMessage(id int64) (*dbmodels.UserMessage, error) {
	data := &dbmodels.UserMessage{}
	err := d.userMessageTable().Find(db.Cond{"id": id}).One(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *UserOperations) ListUserMessages(toUserId int64, afterId int64, limit int) ([]dbmodels.UserMessage, error) {
	messages := make([]dbmodels.UserMessage, 0)
	if limit > 1000 || limit <= 0 {
		limit = 100
	}
	cond := db.Cond{"to_user": toUserId}
	// Use cursor-based pagination: get messages with id less than afterId (since we order by -id)
	if afterId > 0 {
		cond["id <"] = afterId
	}
	err := d.userMessageTable().Find(cond).
		OrderBy("-id").
		Limit(limit).
		All(&messages)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (d *UserOperations) QueryNewMessages(toUserId int64, readHead int64) ([]dbmodels.UserMessage, error) {
	messages := make([]dbmodels.UserMessage, 0)
	err := d.userMessageTable().Find(db.Cond{"to_user": toUserId, "id >": readHead}).
		OrderBy("id ASC").
		All(&messages)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (d *UserOperations) QueryMessageHistory(toUserId int64, limit int) ([]dbmodels.UserMessage, error) {
	messages := make([]dbmodels.UserMessage, 0)
	if limit > 1000 || limit <= 0 {
		limit = 100
	}
	err := d.userMessageTable().Find(db.Cond{"to_user": toUserId}).
		OrderBy("-id").
		Limit(limit).
		All(&messages)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (d *UserOperations) SetMessageAsRead(id int64, toUserId int64) error {
	return d.userMessageTable().Find(db.Cond{"id": id, "to_user": toUserId}).Update(map[string]any{"is_read": true})
}

func (d *UserOperations) SetAllMessagesAsRead(toUserId int64, readHead int64) error {
	// Update user's read head
	err := d.userTable().Find(db.Cond{"id": toUserId}).Update(map[string]any{"msg_read_head": readHead})
	if err != nil {
		return err
	}
	// Mark all messages up to readHead as read
	_, err = d.db.SQL().Exec("UPDATE UserMessages SET is_read = TRUE WHERE to_user = ? AND id <= ?", toUserId, readHead)
	return err
}

func (d *UserOperations) UpdateUserMessage(id int64, data map[string]any) error {
	return d.userMessageTable().Find(db.Cond{"id": id}).Update(data)
}

func (d *UserOperations) DeleteUserMessage(id int64) error {
	return d.userMessageTable().Find(db.Cond{"id": id}).Delete()
}

func (d *UserOperations) userMessageTable() db.Collection {
	return d.db.Collection("UserMessages")
}
