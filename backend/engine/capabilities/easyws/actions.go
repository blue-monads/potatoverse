package easyws

import (
	"encoding/json"
	"errors"

	"github.com/blue-monads/turnix/backend/engine/capabilities/easyws/room"
	"github.com/blue-monads/turnix/backend/xtypes"
)

func (c *EasyWsCapability) Execute(name string, params xtypes.LazyData) (any, error) {
	switch name {
	case "broadcast":
		return c.executeBroadcast(params)
	case "publish":
		return c.executePublish(params)
	case "direct_message":
		return c.executeDirectMessage(params)
	case "subscribe":
		return c.executeSubscribe(params)
	case "unsubscribe":
		return c.executeUnsubscribe(params)
	default:
		return nil, errors.New("unknown action: " + name)
	}
}

func (c *EasyWsCapability) executeBroadcast(params xtypes.LazyData) (any, error) {
	message, err := params.AsBytes()
	if err != nil {
		return nil, err
	}

	err = c.room.Broadcast(message)
	if err != nil {
		return nil, err
	}

	return Ok, nil
}

type PublishParams struct {
	Topic   string          `json:"topic"`
	Message json.RawMessage `json:"message"`
}

func (c *EasyWsCapability) executePublish(params xtypes.LazyData) (any, error) {
	var p PublishParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	if p.Topic == "" {
		return nil, errors.New("topic is required")
	}

	err := c.room.Publish(p.Topic, p.Message)
	if err != nil {
		return nil, err
	}

	return Ok, nil
}

type DirectMessageParams struct {
	TargetConnId string          `json:"target_conn_id"`
	Message      json.RawMessage `json:"message"`
}

func (c *EasyWsCapability) executeDirectMessage(params xtypes.LazyData) (any, error) {
	var p DirectMessageParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	if p.TargetConnId == "" {
		return nil, errors.New("target_conn_id is required")
	}

	err := c.room.DirectMessage(room.ConnId(p.TargetConnId), p.Message)
	if err != nil {
		return nil, err
	}

	return Ok, nil
}

type SubscribeParams struct {
	Topic  string `json:"topic"`
	ConnId string `json:"conn_id"`
}

func (c *EasyWsCapability) executeSubscribe(params xtypes.LazyData) (any, error) {
	var p SubscribeParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	if p.Topic == "" {
		return nil, errors.New("topic is required")
	}

	if p.ConnId == "" {
		return nil, errors.New("conn_id is required")
	}

	err := c.room.Subscribe(p.Topic, room.ConnId(p.ConnId))
	if err != nil {
		return nil, err
	}

	return Ok, nil
}

type UnsubscribeParams struct {
	Topic  string `json:"topic"`
	ConnId string `json:"conn_id"`
}

func (c *EasyWsCapability) executeUnsubscribe(params xtypes.LazyData) (any, error) {
	var p UnsubscribeParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	if p.Topic == "" {
		return nil, errors.New("topic is required")
	}

	if p.ConnId == "" {
		return nil, errors.New("conn_id is required")
	}

	err := c.room.Unsubscribe(p.Topic, room.ConnId(p.ConnId))
	if err != nil {
		return nil, err
	}

	return Ok, nil
}
