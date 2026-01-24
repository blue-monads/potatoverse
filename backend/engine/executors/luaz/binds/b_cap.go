package binds

import (
	"github.com/blue-monads/potatoverse/backend/engine/executors/luaz/lazylua"
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/utils/luaplus"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	lua "github.com/yuin/gopher-lua"
)

func CapBindable(app xtypes.App) map[string]lua.LGFunction {

	capHub := app.Engine().(xtypes.Engine).GetCapabilityHub().(xcapability.CapabilityHub)
	sdb := app.Database().GetSpaceOps()
	s := app.Signer()

	return map[string]lua.LGFunction{
		"list": func(L *lua.LState) int {
			return capList(capHub, L)
		},
		"execute": func(L *lua.LState) int {
			return capExecute(capHub, L)
		},
		"methods": func(L *lua.LState) int {
			return capMethods(capHub, L)
		},
		"sign_token": func(L *lua.LState) int {
			return capSignToken(sdb, s, L)
		},
	}

}

func capList(chub xcapability.CapabilityHub, L *lua.LState) int {
	execState := GetExecState(L)

	caps, err := chub.List(execState.SpaceId)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	resultTable := L.NewTable()
	for _, cap := range caps {
		resultTable.Append(lua.LString(cap))
	}
	L.Push(resultTable)
	return 1
}

func capExecute(chub xcapability.CapabilityHub, L *lua.LState) int {
	execState := GetExecState(L)
	capabilityName := L.CheckString(1)
	method := L.CheckString(2)
	params := L.CheckTable(3)

	paramsLazyData := lazylua.NewLuaLazyData(L, params)
	result, err := chub.Execute(execState.InstalledId, execState.SpaceId, capabilityName, method, paramsLazyData)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	resultTable := luaplus.GoTypeToLuaType(L, result)
	L.Push(resultTable)
	return 1
}

func capMethods(chub xcapability.CapabilityHub, L *lua.LState) int {
	execState := GetExecState(L)
	capabilityName := L.CheckString(1)
	methods, err := chub.Methods(execState.InstalledId, execState.SpaceId, capabilityName)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	resultTable := L.NewTable()
	for _, method := range methods {
		resultTable.Append(lua.LString(method))
	}
	L.Push(resultTable)
	return 1
}

type SignCapabilityTokenOptions struct {
	ResourceId string         `json:"resource_id"`
	ExtraMeta  map[string]any `json:"extrameta"`
	UserId     int64          `json:"user_id"`
}

func capSignToken(sdb datahub.SpaceOps, s *signer.Signer, L *lua.LState) int {
	execState := GetExecState(L)
	capName := L.CheckString(1)
	opts := &SignCapabilityTokenOptions{}
	err := luaplus.MapToStruct(L, L.CheckTable(2), opts)

	capability, err := sdb.GetSpaceCapability(execState.InstalledId, capName)
	if err != nil {
		return luaplus.PushError(L, err)
	}

	signature, err := s.SignCapability(&signer.CapabilityClaim{
		CapabilityId: capability.ID,
		InstallId:    execState.InstalledId,
		SpaceId:      execState.SpaceId,
		UserId:       opts.UserId,
		ResourceId:   opts.ResourceId,
		ExtraMeta:    opts.ExtraMeta,
	})
	if err != nil {
		return luaplus.PushError(L, err)
	}

	L.Push(lua.LString(signature))
	L.Push(lua.LNil)

	return 2
}
