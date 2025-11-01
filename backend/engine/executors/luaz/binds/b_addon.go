package binds

import (
	"github.com/blue-monads/turnix/backend/engine/executors"
	"github.com/blue-monads/turnix/backend/utils/luaplus"
	"github.com/blue-monads/turnix/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

func CapabilityModule(handle *executors.EHandle) func(L *lua.LState) int {
	return func(L *lua.LState) int {

		engine := handle.App.Engine().(xtypes.Engine)
		capabilities := engine.GetCapabilityHub().(xtypes.CapabilityHub)

		listCapabilities := func(L *lua.LState) int {
			caps, err := capabilities.List(handle.SpaceId)
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

		getCapabilityMeta := func(L *lua.LState) int {
			capabilityName := L.CheckString(1)
			method := L.CheckString(2)
			meta, err := capabilities.GetMeta(handle.SpaceId, capabilityName, method)
			if err != nil {
				return pushError(L, err)
			}

			table := luaplus.MapToTable(L, meta)
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
			result, err := capabilities.Execute(handle.SpaceId, capabilityName, method, paramsLazyData)
			if err != nil {
				return pushError(L, err)
			}
			table := luaplus.MapToTable(L, result)
			L.Push(table)
			return 1
		}

		getCapabilityMethods := func(L *lua.LState) int {
			capabilityName := L.CheckString(1)
			methods, err := capabilities.Methods(handle.SpaceId, capabilityName)
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
			"meta":    getCapabilityMeta,
			"execute": executeCapability,
			"methods": getCapabilityMethods,
		})
		L.Push(table)
		return 1
	}
}
