package binds

import (
	"github.com/blue-monads/potatoverse/backend/engine/executors"
	lua "github.com/yuin/gopher-lua"
)

func init() {
	RegisterBindable("db", DBBindable)
	RegisterBindable("kv", KVBindable)
	RegisterBindable("core", CoreBindable)
	RegisterBindable("cap", CapBindable)
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
