package binds

import (
	"encoding/json"
	"strings"

	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/utils/luaplus"
	"github.com/blue-monads/turnix/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

// presign package file presigned token
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
	case "read_package_file":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return readPackageFile(mod, L)
		}))
		return 1
	default:
		return 0
	}

}

type PublishEventOptions struct {
	Name       string `json:"name"`
	Payload    any    `json:"payload"`
	ResourceId string `json:"resource_id"`
}

func corePublishEvent(mod *luaCoreModule, L *lua.LState) int {
	opts := &PublishEventOptions{}
	err := luaplus.MapToStruct(L, L.CheckTable(1), opts)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	var payloadBytes []byte
	if opts.Payload == nil {
		payloadBytes = []byte{}
	} else {
		switch v := opts.Payload.(type) {
		case string:
			payloadBytes = []byte(v)
		case []byte:
			payloadBytes = v
		default:
			// Marshal to JSON for other types (maps, arrays, etc.)
			jsonData, err := json.Marshal(v)
			if err != nil {
				L.Push(lua.LString(err.Error()))
				return 1
			}
			payloadBytes = jsonData
		}
	}

	err = mod.engine.PublishEvent(&xtypes.EventOptions{
		InstallId:  mod.installId,
		Name:       opts.Name,
		Payload:    payloadBytes,
		ResourceId: opts.ResourceId,
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

func readPackageFile(mod *luaCoreModule, L *lua.LState) int {
	fpath := L.CheckString(1)

	fileName := fpath
	dirPath := ""

	if strings.Contains(fpath, "/") {
		parts := strings.Split(fpath, "/")
		fileName = parts[len(parts)-1]
		dirPath = strings.Join(parts[:len(parts)-1], "/")
	}

	pops := mod.app.Database().GetPackageFileOps()
	fileData, err := pops.GetFileContentByPath(mod.installId, dirPath, fileName)
	if err != nil {
		return pushError(L, err)
	}
	L.Push(lua.LString(fileData))
	L.Push(lua.LNil)
	return 2

}
