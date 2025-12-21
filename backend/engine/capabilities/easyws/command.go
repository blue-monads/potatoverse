package easyws

import (
	"fmt"

	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/xcapability/easyaction"
)

func (c *EasyWsCapability) handleCommand() {
	engine := c.builder.app.Engine().(xtypes.Engine)

	for cmd := range c.cmdChan {

		ctx := easyaction.Context{
			Capability: c,
			Payload:    cmd.Data,
		}

		err := engine.EmitActionEvent(&xtypes.ActionEventOptions{
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

		if err != nil {
			qq.Println("@handle_command/1", "error executing action", err)
		}

	}

}
