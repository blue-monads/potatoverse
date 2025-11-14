package binds

import (
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

func CoreModule(app xtypes.App, installId int64, spaceId int64) func(L *lua.LState) int {
	engine := app.Engine().(xtypes.Engine)
	sig := app.Signer()

	return func(L *lua.LState) int {
		publishEvent := func(L *lua.LState) int {
			name := L.CheckString(1)
			payload := L.CheckString(2)
			err := engine.PublishEvent(installId, name, []byte(payload))
			if err != nil {
				L.Push(lua.LString(err.Error()))
				return 1
			}
			L.Push(lua.LNil)
			return 1
		}

		signFsPresignedToken := func(L *lua.LState) int {
			opts := &SignFsPresignedTokenOptions{}
			err := luaplus.MapToStruct(L, L.CheckTable(1), opts)
			if err != nil {
				L.Push(lua.LString(err.Error()))
				return 1
			}

			signature, err := sig.SignSpaceFilePresigned(&signer.SpaceFilePresignedClaim{
				InstallId: installId,
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

		table := L.NewTable()
		L.SetFuncs(table, map[string]lua.LGFunction{
			"publish_event":          publishEvent,
			"get_fs_presigned_token": signFsPresignedToken,
		})

		L.Push(table)
		return 1
	}
}
