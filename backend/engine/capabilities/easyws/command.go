package easyws

import (
	"fmt"

	"github.com/blue-monads/turnix/backend/engine/capabilities/easyws/room"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/xcapability/easyaction"
)

type CMDMessage struct {
	c   *EasyWsCapability
	cmd room.Message
}

type CMDDisconnectMessage struct {
	c     *EasyWsCapability
	uinfo room.UserConnInfo
}

func (c *EasyWsBuilder) evLoop() {

	qq.Println("@evLoop/0")

	for {

		select {
		case cmd := <-c.onCmdChan:

			qq.Println("@evLoop/1", "command", cmd.cmd.Target)

			go func() {

				err := cmd.c.onCommand(cmd.cmd)
				if err != nil {
					qq.Println("@evLoop/1", "error executing command", err)
				}
			}()
		case uinfo := <-c.onDisconnectChan:
			qq.Println("@evLoop/2", "disconnect", uinfo.uinfo.ConnId, uinfo.uinfo.UserId)

			go func() {

				qq.Println("@evLoop/2", "disconnecting", uinfo.uinfo.ConnId, uinfo.uinfo.UserId)

				err := uinfo.c.afterDisconnect(string(uinfo.uinfo.ConnId), uinfo.uinfo.UserId)
				if err != nil {
					qq.Println("@evLoop/2", "error executing disconnect", err)
				}
			}()

		}

	}

}

func (c *EasyWsCapability) onCommand(cmd room.Message) error {
	ctx := easyaction.Context{
		Capability: c,
		Payload:    cmd.Data,
	}

	err := c.builder.engine.EmitActionEvent(&xtypes.ActionEventOptions{
		SpaceId:    c.spaceId,
		EventType:  "capability",
		ActionName: "client_command",
		Params: map[string]string{
			"command":       cmd.Target,
			"capability_id": fmt.Sprintf("%d", c.capabilityId),
			"capability":    "easyws",
		},
		Request: &ctx,
	})

	return err
}

func (c *EasyWsCapability) afterConnect(connId, ip string, userId int64) error {

	qq.Println("afterConnect/1", connId, userId)

	ctx := easyaction.Context{
		Capability: c,
		Payload:    nil,
	}

	err := c.builder.engine.EmitActionEvent(&xtypes.ActionEventOptions{
		SpaceId:    c.spaceId,
		EventType:  "capability",
		ActionName: "after_connect",
		Params: map[string]string{
			"conn_id":       connId,
			"capability_id": fmt.Sprintf("%d", c.capabilityId),
			"capability":    "easyws",
			"user_id":       fmt.Sprintf("%d", userId),
			"ip":            ip,
		},
		Request: &ctx,
	})

	return err
}

func (c *EasyWsCapability) afterDisconnect(connId string, userId int64) error {

	ctx := easyaction.Context{
		Capability: c,
		Payload:    nil,
	}

	err := c.builder.engine.EmitActionEvent(&xtypes.ActionEventOptions{
		SpaceId:    c.spaceId,
		EventType:  "capability",
		ActionName: "after_disconnect",
		Params: map[string]string{
			"conn_id":       connId,
			"capability_id": fmt.Sprintf("%d", c.capabilityId),
			"capability":    "easyws",
			"user_id":       fmt.Sprintf("%d", userId),
		},
		Request: &ctx,
	})

	return err

}
