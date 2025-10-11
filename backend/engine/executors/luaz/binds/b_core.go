package binds

import (
	"github.com/blue-monads/turnix/backend/engine/executors"
	lua "github.com/yuin/gopher-lua"
)

func CoreModule(handle *executors.EHandle) func(L *lua.LState) int {

	return func(L *lua.LState) int {

		GetSpaceFilePresigned := func(L *lua.LState) int {
			uid := L.CheckInt64(1)
			path := L.CheckString(2)
			fileName := L.CheckString(3)
			expiry := L.OptInt64(4, 3600) // Default to 1 hour if not provided

			token, err := handle.GetSpaceFilePresigned(uid, path, fileName, expiry)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(lua.LString(token))
			return 1
		}

		table := L.NewTable()
		L.SetFuncs(table, map[string]lua.LGFunction{
			"get_space_file_presigned": GetSpaceFilePresigned,
		})
		L.Push(table)
		return 1
	}
}
