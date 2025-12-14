package easyws

import "github.com/blue-monads/turnix/backend/xtypes"

func (c *EasyWsCapability) handleCommand() {
	engine := c.builder.app.Engine().(xtypes.Engine)

	for cmd := range c.cmdChan {
		engine.ExecAction(&xtypes.EngineActionExecution{
			SpaceId:    c.spaceId,
			ActionType: "ws_command",
			ActionName: cmd.Target,
			Params:     nil,
			Request:    nil,
		})

	}

}
