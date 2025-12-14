package easyws

import (
	"github.com/blue-monads/turnix/backend/engine/capabilities/easyws/room"
	"github.com/blue-monads/turnix/backend/xtypes"
)

func (c *EasyWsCapability) handleCommand() {
	engine := c.builder.app.Engine().(xtypes.Engine)

	for cmd := range c.cmdChan {
		engine.ExecAction(&xtypes.EngineActionExecution{
			SpaceId:    c.spaceId,
			ActionType: "ws_command",
			ActionName: cmd.Target,
			Params:     nil,
			Request: &ActionContext{
				c:   c,
				cmd: cmd,
			},
		})

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
