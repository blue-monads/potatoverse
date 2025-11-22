package binds

import (
	"encoding/json"

	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/utils/luaplus"
	"github.com/blue-monads/turnix/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

// presign package file presigned token
// package_get_file
// package_list_folder
// write_temporary_file
// read_temporary_file
// delete_temporary_file
// create_temporary_folder
// delete_temporary_folder
// list_temporary_folder
// read_seek_temporary_file
// write_seek_temporary_file
// get_temporary_file_info

type SignFsPresignedTokenOptions struct {
	Path     string `json:"path"`
	FileName string `json:"file_name"`
	UserId   int64  `json:"user_id"`
}

// Core Module
func registerCoreModuleType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaCoreModuleTypeName)
	L.SetField(mt, "__index", L.NewFunction(coreModuleIndex))
}

func newCoreModule(L *lua.LState, app xtypes.App, installId int64, spaceId int64) *lua.LUserData {
	engine := app.Engine().(xtypes.Engine)
	ud := L.NewUserData()
	ud.Value = &luaCoreModule{
		app:       app,
		installId: installId,
		spaceId:   spaceId,
		engine:    engine,
		sig:       app.Signer(),
	}
	L.SetMetatable(ud, L.GetTypeMetatable(luaCoreModuleTypeName))
	return ud
}

func checkCoreModule(L *lua.LState) *luaCoreModule {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*luaCoreModule); ok {
		return v
	}
	L.ArgError(1, luaCoreModuleTypeName+" expected")
	return nil
}

func coreModuleIndex(L *lua.LState) int {
	mod := checkCoreModule(L)
	method := L.CheckString(2)

	switch method {
	case "publish_event":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return corePublishEvent(mod, L)
		}))
		return 1
	case "publish_json_event":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return corePublishJSONEvent(mod, L)
		}))
		return 1
	case "file_token":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return coreSignFsPresignedToken(mod, L)
		}))
		return 1
	case "advisery_token":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return coreSignAdviseryToken(mod, L)
		}))
		return 1
	}

	return 0
}

func corePublishEvent(mod *luaCoreModule, L *lua.LState) int {
	name := L.CheckString(1)
	payload := L.CheckString(2)
	err := mod.engine.PublishEvent(&xtypes.EventOptions{
		InstallId:  mod.installId,
		Name:       name,
		Payload:    []byte(payload),
		ResourceId: "",
	})
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

/*

type PublishJSONEventOptions struct {
	Name       string         `json:"name"`
	Payload    map[string]any `json:"payload"`
	ResourceId string         `json:"resource_id"`
}

*/

func corePublishJSONEvent(mod *luaCoreModule, L *lua.LState) int {
	name := L.CheckString(1)
	payload := L.CheckTable(2)
	payloadMap := luaplus.TableToMap(L, payload)
	jsonData, err := json.Marshal(payloadMap)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	err = mod.engine.PublishEvent(&xtypes.EventOptions{
		InstallId: mod.installId,
		Name:      name,
		Payload:   jsonData,
	})
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

func coreSignFsPresignedToken(mod *luaCoreModule, L *lua.LState) int {
	opts := &SignFsPresignedTokenOptions{}
	err := luaplus.MapToStruct(L, L.CheckTable(1), opts)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	signature, err := mod.sig.SignSpaceFilePresigned(&signer.SpaceFilePresignedClaim{
		InstallId: mod.installId,
		UserId:    opts.UserId,
		PathName:  opts.Path,
		FileName:  opts.FileName,
	})
	if err != nil {
		return pushError(L, err)
	}
	L.Push(lua.LString(signature))
	L.Push(lua.LNil)
	return 2
}

type SignAdviseryTokenOptions struct {
	TokenSubType string         `json:"token_sub_type"`
	UserId       int64          `json:"user_id"`
	Data         map[string]any `json:"data"`
}

func coreSignAdviseryToken(mod *luaCoreModule, L *lua.LState) int {
	opts := &SignAdviseryTokenOptions{}
	err := luaplus.MapToStruct(L, L.CheckTable(1), opts)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	signature, err := mod.sig.SignSpaceAdvisiery(&signer.SpaceAdvisieryClaim{
		InstallId:    mod.installId,
		UserId:       opts.UserId,
		TokenSubType: opts.TokenSubType,
		Data:         opts.Data,
		SpaceId:      mod.spaceId,
	})
	if err != nil {
		return pushError(L, err)
	}
	L.Push(lua.LString(signature))
	L.Push(lua.LNil)
	return 2
}
