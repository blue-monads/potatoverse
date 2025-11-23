package chighsock

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/services/sockd/higher"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
)

type ChighsockCapability struct {
	app          xtypes.App
	spaceId      int64
	installId    int64
	capabilityId int64
	signer       *signer.Signer
	higher       *higher.HigherSockd
}

func (c *ChighsockCapability) Handle(ctx *gin.Context) {
	// Try to get user ID from capability token first
	var userId int64
	var err error
	var connId int64

	token := ctx.Request.Header.Get("x-cap-token")
	if token != "" {
		claim, err := c.signer.ParseCapability(token)
		if err == nil {
			if claim.SpaceId != c.spaceId || claim.InstallId != c.installId || claim.CapabilityId != c.capabilityId {
				httpx.WriteErrString(ctx, "token validation failed")
				return
			}
			userId = claim.UserId

			connId, _ = strconv.ParseInt(claim.ResourceId, 10, 64)
			if connId == 0 {
				httpx.WriteErrString(ctx, "Resource ID is not a valid integer")
				return
			}

		}
	}

	if userId == 0 {
		httpx.WriteErrString(ctx, "authentication required")
		return
	}

	// Get room name from query parameter or use default
	roomName := fmt.Sprintf("cap-%d", c.capabilityId)

	// Upgrade to websocket
	conn, _, _, err := ws.UpgradeHTTP(ctx.Request, ctx.Writer)
	if err != nil {
		httpx.WriteErrString(ctx, "failed to upgrade websocket")
		return
	}

	_, err = c.higher.AddConn(userId, conn, connId, roomName)
	if err != nil {
		conn.Close()
		httpx.WriteErrString(ctx, "failed to add connection")
		return
	}
}

func (c *ChighsockCapability) ListActions() ([]string, error) {
	return []string{
		"broadcast",
		"publish",
		"direct_message",
		"subscribe",
		"unsubscribe",
	}, nil
}

func (c *ChighsockCapability) Execute(name string, params xtypes.LazyData) (map[string]any, error) {
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

func (c *ChighsockCapability) executeBroadcast(params xtypes.LazyData) (map[string]any, error) {
	message, err := params.AsBytes()
	if err != nil {
		return nil, err
	}

	roomName := fmt.Sprintf("cap-%d", c.capabilityId)

	err = c.higher.Broadcast(roomName, message)
	if err != nil {
		return nil, err
	}

	return OKResponse, nil
}

type PublishParams struct {
	Topic   string          `json:"topic"`
	Message json.RawMessage `json:"message"`
}

func (c *ChighsockCapability) executePublish(params xtypes.LazyData) (map[string]any, error) {
	var p PublishParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	roomName := fmt.Sprintf("cap-%d", c.capabilityId)

	if p.Topic == "" {
		return nil, errors.New("topic is required")
	}

	err := c.higher.Publish(roomName, p.Topic, p.Message)
	if err != nil {
		return nil, err
	}

	return OKResponse, nil
}

type DirectMessageParams struct {
	TargetConnId int64           `json:"target_conn_id"`
	Message      json.RawMessage `json:"message"`
}

func (c *ChighsockCapability) executeDirectMessage(params xtypes.LazyData) (map[string]any, error) {
	var p DirectMessageParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	roomName := fmt.Sprintf("cap-%d", c.capabilityId)

	if p.TargetConnId == 0 {
		return nil, errors.New("target_conn_id is required")
	}

	err := c.higher.DirectMessage(roomName, p.TargetConnId, p.Message)
	if err != nil {
		return nil, err
	}

	return OKResponse, nil
}

type SubscribeParams struct {
	Topic  string `json:"topic"`
	ConnId int64  `json:"conn_id"`
}

func (c *ChighsockCapability) executeSubscribe(params xtypes.LazyData) (map[string]any, error) {
	var p SubscribeParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	if p.Topic == "" {
		return nil, errors.New("topic is required")
	}

	if p.ConnId == 0 {
		return nil, errors.New("conn_id is required")
	}

	roomName := fmt.Sprintf("cap-%d", c.capabilityId)

	err := c.higher.Subscribe(roomName, p.Topic, p.ConnId)
	if err != nil {
		return nil, err
	}

	return OKResponse, nil
}

type UnsubscribeParams struct {
	Topic  string `json:"topic"`
	ConnId int64  `json:"conn_id"`
}

func (c *ChighsockCapability) executeUnsubscribe(params xtypes.LazyData) (map[string]any, error) {
	var p UnsubscribeParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	if p.Topic == "" {
		return nil, errors.New("topic is required")
	}

	if p.ConnId == 0 {
		return nil, errors.New("conn_id is required")
	}

	roomName := fmt.Sprintf("cap-%d", c.capabilityId)

	err := c.higher.Unsubscribe(roomName, p.Topic, p.ConnId)
	if err != nil {
		return nil, err
	}

	return OKResponse, nil
}

func (c *ChighsockCapability) Reload(model *dbmodels.SpaceCapability) (xtypes.Capability, error) {
	return &ChighsockCapability{
		app:          c.app,
		spaceId:      model.SpaceID,
		installId:    model.InstallID,
		capabilityId: model.ID,
		signer:       c.signer,
		higher:       c.higher,
	}, nil
}

func (c *ChighsockCapability) Close() error {
	// Cleanup can be done here if needed
	return nil
}
