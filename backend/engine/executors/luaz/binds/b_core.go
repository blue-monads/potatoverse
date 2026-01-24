package binds

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/blue-monads/potatoverse/backend/engine/executors"
	"github.com/blue-monads/potatoverse/backend/services/corehub"
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/utils/luaplus"
	"github.com/blue-monads/potatoverse/backend/xtypes"
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

func CoreBindable(app xtypes.App) map[string]lua.LGFunction {

	engine := app.Engine().(xtypes.Engine)
	sig := app.Signer()
	pops := app.Database().GetPackageFileOps()

	return map[string]lua.LGFunction{
		"publish_event": func(L *lua.LState) int {
			return corePublishEvent(engine, GetExecState(L), L)
		},
		"file_token": func(L *lua.LState) int {
			return coreSignFsPresignedToken(sig, GetExecState(L), L)
		},
		"sign_advisery_token": func(L *lua.LState) int {
			return coreSignAdviseryToken(sig, GetExecState(L), L)
		},
		"parse_advisery_token": func(L *lua.LState) int {
			return coreParseAdviseryToken(sig, GetExecState(L), L)
		},
		"read_package_file": func(L *lua.LState) int {
			return readPackageFile(pops, GetExecState(L), L)
		},
		"list_files": func(L *lua.LState) int {
			coreHub := app.CoreHub().(*corehub.CoreHub)

			return coreListFiles(coreHub, GetExecState(L), L)
		},
		"decode_file_id": func(L *lua.LState) int {
			coreHub := app.CoreHub().(*corehub.CoreHub)
			return coreDecodeFileId(coreHub, L)
		},
		"encode_file_id": func(L *lua.LState) int {
			coreHub := app.CoreHub().(*corehub.CoreHub)
			return coreEncodeFileId(coreHub, L)
		},
		"db_vendor": func(L *lua.LState) int {
			L.Push(lua.LString(app.Database().Vender()))
			return 1
		},
	}

}

type PublishEventOptions struct {
	Name        string `json:"name"`
	Payload     any    `json:"payload"`
	ResourceId  string `json:"resource_id"`
	CollapseKey string `json:"collapse_key"`
}

func corePublishEvent(engine xtypes.Engine, es *executors.ExecState, L *lua.LState) int {
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

	err = engine.PublishEvent(&xtypes.EventOptions{
		InstallId:   es.InstalledId,
		Name:        opts.Name,
		Payload:     payloadBytes,
		ResourceId:  opts.ResourceId,
		CollapseKey: opts.CollapseKey,
		SpaceId:     es.SpaceId,
	})
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

type SignFsPresignedTokenOptions struct {
	Path     string `json:"path"`
	FileName string `json:"file_name"`
	UserId   int64  `json:"user_id"`
}

func coreSignFsPresignedToken(sig *signer.Signer, es *executors.ExecState, L *lua.LState) int {
	opts := &SignFsPresignedTokenOptions{}
	err := luaplus.MapToStruct(L, L.CheckTable(1), opts)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	signature, err := sig.SignSpaceFilePresigned(&signer.SpaceFilePresignedClaim{
		InstallId: es.InstalledId,
		UserId:    opts.UserId,
		PathName:  opts.Path,
		FileName:  opts.FileName,
	})
	if err != nil {
		return luaplus.PushError(L, err)
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

func coreSignAdviseryToken(sig *signer.Signer, es *executors.ExecState, L *lua.LState) int {
	opts := &SignAdviseryTokenOptions{}
	err := luaplus.MapToStruct(L, L.CheckTable(1), opts)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	signature, err := sig.SignSpaceAdvisiery(&signer.SpaceAdvisieryClaim{
		InstallId:    es.InstalledId,
		UserId:       opts.UserId,
		TokenSubType: opts.TokenSubType,
		Data:         opts.Data,
		SpaceId:      es.SpaceId,
	})
	if err != nil {
		return luaplus.PushError(L, err)
	}
	L.Push(lua.LString(signature))
	L.Push(lua.LNil)
	return 2
}

func coreParseAdviseryToken(sig *signer.Signer, es *executors.ExecState, L *lua.LState) int {
	token := L.CheckString(1)
	claim, err := sig.ParseSpaceAdvisiery(token)
	if err != nil {
		return luaplus.PushError(L, err)
	}

	if claim.InstallId != es.InstalledId {
		return luaplus.PushError(L, errors.New("wrong install id"))
	}

	if claim.SpaceId != es.SpaceId {
		return luaplus.PushError(L, errors.New("wrong space id"))
	}

	resultTable := L.NewTable()

	resultTable.RawSetString("token_sub_type", lua.LString(claim.TokenSubType))
	resultTable.RawSetString("user_id", lua.LNumber(claim.UserId))
	resultTable.RawSetString("data", luaplus.MapToTable(L, claim.Data))

	L.Push(resultTable)
	L.Push(lua.LNil)
	return 2
}

func readPackageFile(pops datahub.FileOps, es *executors.ExecState, L *lua.LState) int {
	fpath := L.CheckString(1)

	fileName := fpath
	dirPath := ""

	if strings.Contains(fpath, "/") {
		parts := strings.Split(fpath, "/")
		fileName = parts[len(parts)-1]
		dirPath = strings.Join(parts[:len(parts)-1], "/")
	}

	fileData, err := pops.GetFileContentByPath(es.PackageVersionId, dirPath, fileName)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	L.Push(lua.LString(fileData))
	L.Push(lua.LNil)
	return 2

}

func coreListFiles(fops *corehub.CoreHub, es *executors.ExecState, L *lua.LState) int {
	path := L.OptString(1, "")
	files, err := fops.ListSpaceFilesSigned(es.InstalledId, path)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	resultTable := L.NewTable()
	for _, file := range files {

		fileTable := L.NewTable()
		fileTable.RawSetString("id", lua.LString(file.Id))
		fileTable.RawSetString("name", lua.LString(file.Name))
		fileTable.RawSetString("is_folder", lua.LBool(file.IsFolder))
		fileTable.RawSetString("path", lua.LString(file.Path))
		fileTable.RawSetString("size", lua.LNumber(file.Size))
		fileTable.RawSetString("mime", lua.LString(file.Mime))
		fileTable.RawSetString("hash", lua.LString(file.Hash))
		resultTable.Append(fileTable)
	}

	L.Push(resultTable)
	L.Push(lua.LNil)
	return 2
}

func coreDecodeFileId(coreHub *corehub.CoreHub, L *lua.LState) int {
	id := L.CheckString(1)
	decodedId, err := coreHub.DecodeSpaceFileId(id)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	L.Push(lua.LNumber(decodedId))
	L.Push(lua.LNil)
	return 2
}

func coreEncodeFileId(coreHub *corehub.CoreHub, L *lua.LState) int {
	fid := L.CheckNumber(1)
	encodedId, err := coreHub.EncodeSpaceFileId(int64(fid))
	if err != nil {
		return luaplus.PushError(L, err)
	}
	L.Push(lua.LString(encodedId))
	L.Push(lua.LNil)
	return 2
}
