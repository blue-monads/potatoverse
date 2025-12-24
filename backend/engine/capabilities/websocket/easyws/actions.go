package easyws

import (
	"encoding/json"
	"errors"

	"github.com/blue-monads/turnix/backend/engine/capabilities/websocket/easyws/room"
	"github.com/blue-monads/turnix/backend/xtypes/lazydata"
)

func (c *EasyWsCapability) Execute(name string, params lazydata.LazyData) (any, error) {
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

func (c *EasyWsCapability) executeBroadcast(params lazydata.LazyData) (any, error) {
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

func (c *EasyWsCapability) executePublish(params lazydata.LazyData) (any, error) {
	var p MessageParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	if p.Target == "" {
		return nil, errors.New("topic is required")
	}

	err := c.room.Publish(p.Target, p.Message)
	if err != nil {
		return nil, err
	}

	return Ok, nil
}

type MessageParams struct {
	Target  string          `json:"target"`
	Message json.RawMessage `json:"message"`
}

func (c *EasyWsCapability) executeDirectMessage(params lazydata.LazyData) (any, error) {
	var p MessageParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	if p.Target == "" {
		return nil, errors.New("target_conn_id is required")
	}

	err := c.room.DirectMessage(room.ConnId(p.Target), p.Message)
	if err != nil {
		return nil, err
	}

	return Ok, nil
}

type SubscribeParams struct {
	Topic  string `json:"topic"`
	ConnId string `json:"conn_id"`
}

func (c *EasyWsCapability) executeSubscribe(params lazydata.LazyData) (any, error) {
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

func (c *EasyWsCapability) executeUnsubscribe(params lazydata.LazyData) (any, error) {
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

	err := c.room.Unsubscribe(p.Topic, room.ConnId(p.ConnId))
	if err != nil {
		return nil, err
	}

	return Ok, nil
}
