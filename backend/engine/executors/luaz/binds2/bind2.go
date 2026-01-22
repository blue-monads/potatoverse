package binds2

import (
	"maps"
	"sync"

	"github.com/blue-monads/potatoverse/backend/engine/executors"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

type LuaBindable func(app xtypes.App) map[string]lua.LGFunction

var (
	Bindables     = make(map[string]LuaBindable)
	BindablesLock = sync.Mutex{}
)

func RegisterBindable(name string, bindable LuaBindable) {
	BindablesLock.Lock()
	defer BindablesLock.Unlock()
	Bindables[name] = bindable
}

func GetBindables() map[string]LuaBindable {
	BindablesLock.Lock()
	defer BindablesLock.Unlock()

	return maps.Clone(Bindables)
}

func GetExecState(L *lua.LState) *executors.ExecState {
	execState := L.CheckUserData(1)
	return execState.Value.(*executors.ExecState)
}

const (
	luaPotatoBindableTypeName = "potato.module"
)

func RegisterPotatoBindableType(L *lua.LState, methods map[string]lua.LGFunction) {
	mt := L.NewTypeMetatable(luaPotatoBindableTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))
}
