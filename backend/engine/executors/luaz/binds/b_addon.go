package binds

import (
	"github.com/blue-monads/turnix/backend/engine/executors"
	"github.com/blue-monads/turnix/backend/utils/luaplus"
	"github.com/blue-monads/turnix/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

func AddOnModule(handle *executors.EHandle) func(L *lua.LState) int {
	return func(L *lua.LState) int {

		engine := handle.App.Engine().(xtypes.Engine)
		addons := engine.GetAddonHub().(xtypes.AddOnHub)

		listAddons := func(L *lua.LState) int {
			addons, err := addons.List(handle.SpaceId)
			if err != nil {
				return pushError(L, err)
			}
			table := L.NewTable()
			for _, addon := range addons {
				table.Append(lua.LString(addon))
			}
			L.Push(table)
			return 1
		}

		getAddonMeta := func(L *lua.LState) int {
			addOnName := L.CheckString(1)
			method := L.CheckString(2)
			meta, err := addons.GetMeta(handle.SpaceId, addOnName, method)
			if err != nil {
				return pushError(L, err)
			}

			table := luaplus.MapToTable(L, meta)
			L.Push(table)
			return 1
		}

		executeAddon := func(L *lua.LState) int {
			addOnName := L.CheckString(1)
			method := L.CheckString(2)
			params := L.CheckTable(3)
			paramsLazyData := &LuaLazyData{
				L:     L,
				table: params,
			}
			result, err := addons.Execute(handle.SpaceId, addOnName, method, paramsLazyData)
			if err != nil {
				return pushError(L, err)
			}
			table := luaplus.MapToTable(L, result)
			L.Push(table)
			return 1
		}

		getAddonMethods := func(L *lua.LState) int {
			addOnName := L.CheckString(1)
			methods, err := addons.Methods(handle.SpaceId, addOnName)
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
			"list":    listAddons,
			"meta":    getAddonMeta,
			"execute": executeAddon,
			"methods": getAddonMethods,
		})
		L.Push(table)
		return 1
	}
}
