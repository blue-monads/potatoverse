package binds

import (
	"github.com/blue-monads/turnix/backend/utils/luaplus"
	"github.com/blue-monads/turnix/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

func CapabilityModule(app xtypes.App, installId int64, spaceId int64) func(L *lua.LState) int {
	return func(L *lua.LState) int {

		engine := app.Engine().(xtypes.Engine)
		capabilities := engine.GetCapabilityHub().(xtypes.CapabilityHub)

		listCapabilities := func(L *lua.LState) int {
			caps, err := capabilities.List(spaceId)
			if err != nil {
				return pushError(L, err)
			}
			table := L.NewTable()
			for _, cap := range caps {
				table.Append(lua.LString(cap))
			}
			L.Push(table)
			return 1
		}

		executeCapability := func(L *lua.LState) int {
			capabilityName := L.CheckString(1)
			method := L.CheckString(2)
			params := L.CheckTable(3)
			paramsLazyData := &LuaLazyData{
				L:     L,
				table: params,
			}
			result, err := capabilities.Execute(spaceId, capabilityName, method, paramsLazyData)
			if err != nil {
				return pushError(L, err)
			}
			table := luaplus.MapToTable(L, result)
			L.Push(table)
			return 1
		}

		getCapabilityMethods := func(L *lua.LState) int {
			capabilityName := L.CheckString(1)
			methods, err := capabilities.Methods(spaceId, capabilityName)
			if err != nil {
				return pushError(L, err)
			}
			table := L.NewTable()
			for _, method := range methods {
				table.Append(lua.LString(method))
			}
			L.Push(table)
			return 1
		}

		table := L.NewTable()
		L.SetFuncs(table, map[string]lua.LGFunction{
			"list":    listCapabilities,
			"execute": executeCapability,
			"methods": getCapabilityMethods,
		})
		L.Push(table)
		return 1
	}
}
