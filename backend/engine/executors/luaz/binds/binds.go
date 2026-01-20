package binds

import (
	"github.com/blue-monads/potatoverse/backend/utils/luaplus"
	lua "github.com/yuin/gopher-lua"
)

type HostHandle interface {
	AddCloser(closer func() error) uint16
	RemoveCloser(id uint16)
}

type LuaLazyData struct {
	table *lua.LTable
	L     *lua.LState
}

func pushError(L *lua.LState, err error) int {
	return luaplus.PushError(L, err)
}
