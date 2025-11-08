package luaplus

import lua "github.com/yuin/gopher-lua"

func PushError(L *lua.LState, err error) int {
	L.Push(lua.LNil)
	L.Push(lua.LString(err.Error()))
	return 2
}
