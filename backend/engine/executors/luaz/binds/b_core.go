package binds

import (
	"github.com/blue-monads/turnix/backend/engine/executors"
	lua "github.com/yuin/gopher-lua"
)

/*

AddOn Bindings
 - addons_list
 - addons_action_meta
 - addons_execute_action
 - addons_actions


*/

func AddOnModule(handle *executors.EHandle) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		table := L.NewTable()
		L.SetFuncs(table, map[string]lua.LGFunction{})
		L.Push(table)
		return 1
	}
}
