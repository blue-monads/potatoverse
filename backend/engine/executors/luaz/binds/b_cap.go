package binds

import (
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/utils/luaplus"
	"github.com/blue-monads/turnix/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

type SignCapabilityTokenOptions struct {
	ResourceId string         `json:"resource_id"`
	ExtraMeta  map[string]any `json:"extrameta"`
	UserId     int64          `json:"user_id"`
}

// sign capability token

func CapabilityModule(app xtypes.App, installId int64, spaceId int64) func(L *lua.LState) int {
	return func(L *lua.LState) int {

		engine := app.Engine().(xtypes.Engine)
		capabilities := engine.GetCapabilityHub().(xtypes.CapabilityHub)

		listCapabilities := func(L *lua.LState) int {
			caps, err := capabilities.List(spaceId)
			if err != nil {
				return pushError(L, err)
			}
			table := L.NewTable()
			for _, cap := range caps {
				table.Append(lua.LString(cap))
			}
			L.Push(table)
			return 1
		}

		executeCapability := func(L *lua.LState) int {
			capabilityName := L.CheckString(1)
			method := L.CheckString(2)
			params := L.CheckTable(3)
			paramsLazyData := &LuaLazyData{
				L:     L,
				table: params,
			}
			result, err := capabilities.Execute(spaceId, capabilityName, method, paramsLazyData)
			if err != nil {
				return pushError(L, err)
			}
			table := luaplus.MapToTable(L, result)
			L.Push(table)
			return 1
		}

		getCapabilityMethods := func(L *lua.LState) int {
			capabilityName := L.CheckString(1)
			methods, err := capabilities.Methods(spaceId, capabilityName)
			if err != nil {
				return pushError(L, err)
			}
			table := L.NewTable()
			for _, method := range methods {
				table.Append(lua.LString(method))
			}
			L.Push(table)
			return 1
		}

		signCapabilityToken := func(L *lua.LState) int {
			capName := L.CheckString(1)
			opts := &SignCapabilityTokenOptions{}
			err := luaplus.MapToStruct(L, L.CheckTable(2), opts)
			if err != nil {
				return pushError(L, err)
			}

			sdb := app.Database().GetSpaceOps()

			capability, err := sdb.GetSpaceCapability(installId, capName)
			if err != nil {
				return pushError(L, err)
			}

			s := app.Signer()

			signature, err := s.SignCapability(&signer.CapabilityClaim{
				CapabilityId: capability.ID,
				InstallId:    installId,
				SpaceId:      spaceId,
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

		table := L.NewTable()
		L.SetFuncs(table, map[string]lua.LGFunction{
			"list":       listCapabilities,
			"execute":    executeCapability,
			"methods":    getCapabilityMethods,
			"sign_token": signCapabilityToken,
		})
		L.Push(table)
		return 1
	}
}
