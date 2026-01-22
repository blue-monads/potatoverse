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
	esud := L.GetGlobal("__es__")

	if esud.Type() != lua.LTUserData {
		panic("__es__ is not a userdata")
	}

	udata, ok := esud.(*lua.LUserData)
	if !ok {
		panic("__es__ is not a userdata")
	}

	es, ok := udata.Value.(*executors.ExecState)
	if !ok {
		panic("__es__ is not an executors.ExecState")
	}

	return es
}

const (
	luaPotatoBindableTypeName = "potato.module"
)

func RegisterPotatoBindableType(L *lua.LState, methods map[string]lua.LGFunction) {
	mt := L.NewTypeMetatable(luaPotatoBindableTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))
}
