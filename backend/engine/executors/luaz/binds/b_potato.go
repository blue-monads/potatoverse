package binds

import (
	"github.com/blue-monads/potatoverse/backend/engine/executors"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

const (
	luaPotatoModuleTypeName = "potato.module"
)

type luaPotatoModule struct {
	es     *executors.ExecState
	submod map[string]map[string]lua.LGFunction
}

func registerPotatoModuleType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaPotatoModuleTypeName)
	L.SetField(mt, "__index", L.NewFunction(potatoModuleIndex))
}

func potatoModuleIndex(L *lua.LState) int {

	ppmod := L.CheckUserData(1)

	pmod, ok := ppmod.Value.(*luaPotatoModule)
	if !ok {
		L.ArgError(1, luaPotatoModuleTypeName+" expected")
		return 0
	}

	method := L.CheckString(2)

	table := L.NewTable()
	L.SetFuncs(table, pmod.submod[method])
	L.Push(table)
	return 1
}

func PotatoBindable(app xtypes.App) map[string]map[string]lua.LGFunction {

	submod := make(map[string]map[string]lua.LGFunction)
	for name, bindable := range GetBindables() {
		submod[name] = bindable(app)
	}

	return submod

}

func PotatoModule(es *executors.ExecState, submod map[string]map[string]lua.LGFunction) func(L *lua.LState) int {

	return func(L *lua.LState) int {

		// Register Potato Module Type

		registerPotatoModuleType(L)
		ud := L.NewUserData()
		ud.Value = &luaPotatoModule{
			es:     es,
			submod: submod,
		}
		L.SetMetatable(ud, L.GetTypeMetatable(luaPotatoModuleTypeName))
		L.Push(ud)

		// attach exec state in global
		esud := L.NewUserData()
		esud.Value = es
		L.SetGlobal("__es__", esud)

		return 1
	}
}
