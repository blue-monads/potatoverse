package binds

import (
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/utils/luaplus"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/xcapability"
	lua "github.com/yuin/gopher-lua"
)

type SignCapabilityTokenOptions struct {
	ResourceId string         `json:"resource_id"`
	ExtraMeta  map[string]any `json:"extrameta"`
	UserId     int64          `json:"user_id"`
}

// Cap Module
func registerCapModuleType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaCapModuleTypeName)
	L.SetField(mt, "__index", L.NewFunction(capModuleIndex))
}

func newCapModule(L *lua.LState, app xtypes.App, installId int64, spaceId int64) *lua.LUserData {
	engine := app.Engine().(xtypes.Engine)
	ud := L.NewUserData()
	ud.Value = &luaCapModule{
		app:          app,
		installId:    installId,
		spaceId:      spaceId,
		capabilities: engine.GetCapabilityHub().(xcapability.CapabilityHub),
	}
	L.SetMetatable(ud, L.GetTypeMetatable(luaCapModuleTypeName))
	return ud
}

func checkCapModule(L *lua.LState) *luaCapModule {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*luaCapModule); ok {
		return v
	}
	L.ArgError(1, luaCapModuleTypeName+" expected")
	return nil
}

func capModuleIndex(L *lua.LState) int {
	mod := checkCapModule(L)
	method := L.CheckString(2)

	switch method {
	case "list":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return capList(mod, L)
		}))
		return 1
	case "execute":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return capExecute(mod, L)
		}))
		return 1
	case "list_methods":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return capMethods(mod, L)
		}))
		return 1
	case "sign_token":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return capSignToken(mod, L)
		}))
		return 1
	}

	return 0
}

func capList(mod *luaCapModule, L *lua.LState) int {
	caps, err := mod.capabilities.List(mod.spaceId)
	if err != nil {
		return pushError(L, err)
	}
	resultTable := L.NewTable()
	for _, cap := range caps {
		resultTable.Append(lua.LString(cap))
	}
	L.Push(resultTable)
	return 1
}

func capExecute(mod *luaCapModule, L *lua.LState) int {
	capabilityName := L.CheckString(1)
	method := L.CheckString(2)
	params := L.CheckTable(3)
	paramsLazyData := &LuaLazyData{
		L:     L,
		table: params,
	}
	result, err := mod.capabilities.Execute(mod.installId, mod.spaceId, capabilityName, method, paramsLazyData)
	if err != nil {
		return pushError(L, err)
	}
	resultTable := luaplus.GoTypeToLuaType(L, result)
	L.Push(resultTable)
	return 1
}

func capMethods(mod *luaCapModule, L *lua.LState) int {
	capabilityName := L.CheckString(1)
	methods, err := mod.capabilities.Methods(mod.installId, mod.spaceId, capabilityName)
	if err != nil {
		return pushError(L, err)
	}
	resultTable := L.NewTable()
	for _, method := range methods {
		resultTable.Append(lua.LString(method))
	}
	L.Push(resultTable)
	return 1
}

func capSignToken(mod *luaCapModule, L *lua.LState) int {
	capName := L.CheckString(1)
	opts := &SignCapabilityTokenOptions{}
	err := luaplus.MapToStruct(L, L.CheckTable(2), opts)
	if err != nil {
		return pushError(L, err)
	}

	sdb := mod.app.Database().GetSpaceOps()

	capability, err := sdb.GetSpaceCapability(mod.installId, capName)
	if err != nil {
		return pushError(L, err)
	}

	s := mod.app.Signer()

	signature, err := s.SignCapability(&signer.CapabilityClaim{
		CapabilityId: capability.ID,
		InstallId:    mod.installId,
		SpaceId:      mod.spaceId,
		UserId:       opts.UserId,
		ResourceId:   opts.ResourceId,
		ExtraMeta:    opts.ExtraMeta,
	})
	if err != nil {
		return pushError(L, err)
	}

	L.Push(lua.LString(signature))
	L.Push(lua.LNil)

	return 2
}
