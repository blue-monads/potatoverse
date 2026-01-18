package binds2

import (
	"github.com/blue-monads/potatoverse/backend/utils/luaplus"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	lua "github.com/yuin/gopher-lua"
)

func RegisterModules(app xtypes.App) map[string]lua.LGFunction {

	capabilityHub := app.Engine().(xtypes.Engine).GetCapabilityHub().(xcapability.CapabilityHub)

	listCapabilities := func(L *lua.LState) int {
		execState := GetExecState(L)
		caps, err := capabilityHub.List(execState.SpaceId)
		if err != nil {
			return luaplus.PushError(L, err)
		}
		resultTable := L.NewTable()
		for _, cap := range caps {
			resultTable.Append(lua.LString(cap))
		}
		L.Push(resultTable)
		return 1

	}

	return map[string]lua.LGFunction{
		"cap": listCapabilities,
	}
}
