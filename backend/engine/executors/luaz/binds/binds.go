package binds

import (
	"github.com/blue-monads/turnix/backend/utils/luaplus"
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

func (l *LuaLazyData) AsMap() (map[string]any, error) {
	data := luaplus.TableToMap(l.L, l.table)
	return data, nil
}

func (l *LuaLazyData) AsJson(target any) error {

	err := luaplus.MapToStruct(l.L, l.table, target)
	if err != nil {
		return err
	}

	return nil
}

func pushError(L *lua.LState, err error) int {
	return luaplus.PushError(L, err)
}
