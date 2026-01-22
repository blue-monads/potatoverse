package binds2

import (
	"github.com/blue-monads/potatoverse/backend/engine/executors"
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

func PotatoModule(es *executors.ExecState) func(L *lua.LState) int {

	bindables := GetBindables()
	submod := make(map[string]map[string]lua.LGFunction)
	for name, bindable := range bindables {
		submod[name] = bindable(es.App)
	}

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
