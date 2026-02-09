package binds

import (
	"github.com/blue-monads/potatoverse/backend/utils/luaplus"
	lua "github.com/yuin/gopher-lua"
)

func pushError(L *lua.LState, err error) int {
	return luaplus.PushError(L, err)
}
