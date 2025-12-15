package easyws

import (
	"github.com/blue-monads/turnix/backend/engine/capabilities/easyws/room"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
)

func (c *EasyWsCapability) handleCommand() {
	engine := c.builder.app.Engine().(xtypes.Engine)

	for cmd := range c.cmdChan {
		err := engine.EmitActionEvent(&xtypes.ActionEventOptions{
			SpaceId:    c.spaceId,
			EventType:  "ws_callback",
			ActionName: "client_command",
			Params: map[string]string{
				"command": cmd.Target,
			},
			Request: &ActionContext{c: c, cmd: cmd},
		})

		if err != nil {
			qq.Println("@handle_command/1", "error executing action", err)
		}

	}

}

type ActionContext struct {
	c   *EasyWsCapability
	cmd room.Message
}

func (c *ActionContext) ListActions() ([]string, error) {
	return c.c.ListActions()
}

func (c *ActionContext) ExecuteAction(name string, params xtypes.LazyData) (any, error) {
	return c.c.Execute(name, params)
}
