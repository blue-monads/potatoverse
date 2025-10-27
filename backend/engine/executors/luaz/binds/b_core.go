package binds

import (
	"github.com/blue-monads/turnix/backend/engine/executors"
	"github.com/blue-monads/turnix/backend/utils/luaplus"
	lua "github.com/yuin/gopher-lua"
)

/*

Goodies Bindings
 - goodies_list
 - goodies_action_meta
 - goodies_execute_action
 - goodies_actions


*/

func CoreModule(handle *executors.EHandle) func(L *lua.LState) int {

	return func(L *lua.LState) int {

		GetSpaceFilePresigned := func(L *lua.LState) int {
			opts := executors.PresignedOptions{}
			err := luaplus.MapToStruct(L, L.CheckTable(1), &opts)
			if err != nil {
				return pushError(L, err)
			}

			token := "fixme"

			L.Push(lua.LString(token))
			return 1
		}

		table := L.NewTable()

		L.SetFuncs(table, map[string]lua.LGFunction{
			"file_presigned": GetSpaceFilePresigned,
		})
		L.Push(table)
		return 1
	}
}
