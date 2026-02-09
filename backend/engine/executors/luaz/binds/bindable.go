package binds

import (
	"maps"
	"sync"

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
