package binds

import (
	"github.com/blue-monads/turnix/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

func CoreModule(app xtypes.App, installId int64, spaceId int64) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		return 0
	}
}
