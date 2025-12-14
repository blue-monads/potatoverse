package binds

import (
	"errors"

	"github.com/blue-monads/turnix/backend/utils/luaplus"
	"github.com/blue-monads/turnix/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

const (
	luaGenericContextTypeName = "generic.context"
)

type luaGenericContext struct {
	app            xtypes.App
	installId      int64
	spaceId        int64
	genericRequest xtypes.ActionRequest
}

// Generic Context
func registerGenericContextType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaGenericContextTypeName)
	L.SetField(mt, "__index", L.NewFunction(genericContextIndex))
}

func NewGenericContext(L *lua.LState, app xtypes.App, installId, spaceId int64, genericRequest xtypes.ActionRequest) *lua.LUserData {
	registerGenericContextType(L)
	mt := L.GetTypeMetatable(luaGenericContextTypeName)

	ud := L.NewUserData()
	ud.Value = &luaGenericContext{
		app:            app,
		installId:      installId,
		spaceId:        spaceId,
		genericRequest: genericRequest,
	}
	L.SetMetatable(ud, mt)
	return ud
}

func checkGenericContext(L *lua.LState) *luaGenericContext {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*luaGenericContext); ok {
		return v
	}
	L.ArgError(1, luaGenericContextTypeName+" expected")
	return nil
}

func genericContextIndex(L *lua.LState) int {
	genCtx := checkGenericContext(L)
	method := L.CheckString(2)

	switch method {
	case "list_actions":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return genCtxListActions(genCtx, L)
		}))
		return 1
	case "execute_action":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return genCtxExecuteAction(genCtx, L)
		}))
		return 1
	}

	return 0
}

func genCtxListActions(genCtx *luaGenericContext, L *lua.LState) int {
	if genCtx.genericRequest == nil {
		return pushError(L, errors.New("generic context is nil"))
	}

	actions, err := genCtx.genericRequest.ListActions()
	if err != nil {
		return pushError(L, err)
	}

	table := L.NewTable()
	for i, action := range actions {
		L.RawSetInt(table, i+1, lua.LString(action))
	}
	L.Push(table)
	return 1
}

func genCtxExecuteAction(genCtx *luaGenericContext, L *lua.LState) int {
	if genCtx.genericRequest == nil {
		return pushError(L, errors.New("generic context is nil"))
	}

	actionName := L.CheckString(1)
	paramsTable := L.CheckTable(2)

	// Convert Lua table to LazyData
	lazyData := &LuaLazyData{
		table: paramsTable,
		L:     L,
	}

	result, err := genCtx.genericRequest.ExecuteAction(actionName, lazyData)
	if err != nil {
		return pushError(L, err)
	}

	// Convert result map to Lua table
	resultTable := luaplus.MapToTable(L, result)
	L.Push(resultTable)
	return 1
}

func GenericContextModule(app xtypes.App, installId, spaceId int64, L *lua.LState, genericContext xtypes.ActionRequest) *lua.LUserData {
	return NewGenericContext(L, app, installId, spaceId, genericContext)
}
