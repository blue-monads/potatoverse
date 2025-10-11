package binds

import (
	"bytes"
	"reflect"

	"github.com/blue-monads/turnix/backend/engine/executors"
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/utils/kosher"
	lua "github.com/yuin/gopher-lua"
)

func bindsFS(spaceId int64, db datahub.SpaceFileOps) func(L *lua.LState) int {

	return func(L *lua.LState) int {

		AddFile := func(L *lua.LState) int {
			// Get user ID and file data from Lua
			uid := L.CheckInt64(1)
			path := L.CheckString(2)
			name := L.CheckString(3)
			contentStr := L.CheckString(4)

			reader := bytes.NewReader(kosher.Byte(contentStr))
			id, err := db.StreamAddSpaceFile(spaceId, uid, path, name, reader)

			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(lua.LNumber(id))
			return 1
		}

		AddFolder := func(L *lua.LState) int {
			uid := L.CheckInt64(1)
			path := L.CheckString(2)
			name := L.CheckString(3)

			id, err := db.AddSpaceFolder(spaceId, uid, path, name)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(lua.LNumber(id))
			return 1
		}

		GetFileMeta := func(L *lua.LState) int {
			id := L.CheckInt64(1)

			file, err := db.GetSpaceFileMetaById(id)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(ToTableFromStruct(L, reflect.ValueOf(file)))
			return 1
		}

		ListFilesBySpace := func(L *lua.LState) int {
			path := L.CheckString(1)

			files, err := db.ListSpaceFiles(spaceId, path)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			result := L.NewTable()
			for _, file := range files {
				result.Append(ToTableFromStruct(L, reflect.ValueOf(file)))
			}

			L.Push(result)
			return 1
		}

		RemoveFile := func(L *lua.LState) int {
			id := L.CheckInt64(1)

			err := db.RemoveSpaceFile(spaceId, id)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(lua.LTrue)
			return 1
		}

		UpdateFile := func(L *lua.LState) int {
			id := L.CheckInt64(1)
			data := L.CheckTable(2)
			dataMap := TableToMap(L, data)

			err := db.UpdateSpaceFile(spaceId, id, dataMap)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(lua.LTrue)
			return 1
		}

		AddFileShare := func(L *lua.LState) int {
			fileId := L.CheckInt64(1)
			userId := L.CheckInt64(2)

			shareId, err := db.AddFileShare(fileId, userId, spaceId)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(lua.LString(shareId))
			return 1
		}

		ListFileShares := func(L *lua.LState) int {
			fileId := L.CheckInt64(1)

			shares, err := db.ListFileShares(fileId)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			result := L.NewTable()
			for _, share := range shares {
				result.Append(ToTableFromStruct(L, reflect.ValueOf(share)))
			}

			L.Push(result)
			return 1
		}

		RemoveFileShare := func(L *lua.LState) int {
			userId := L.CheckInt64(1)
			shareId := L.CheckString(2)

			err := db.RemoveFileShare(userId, shareId)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(lua.LTrue)
			return 1
		}

		table := L.NewTable()
		L.SetFuncs(table, map[string]lua.LGFunction{
			"add":          AddFile,
			"add_folder":   AddFolder,
			"get_meta":     GetFileMeta,
			"list":         ListFilesBySpace,
			"remove":       RemoveFile,
			"update":       UpdateFile,
			"add_share":    AddFileShare,
			"list_shares":  ListFileShares,
			"remove_share": RemoveFileShare,
			// "get_presigned_url": ,
		})
		L.Push(table)
		return 1
	}
}

func BindsFS(handle *executors.EHandle) func(L *lua.LState) int {
	return bindsFS(handle.SpaceId, handle.App.Database())
}
