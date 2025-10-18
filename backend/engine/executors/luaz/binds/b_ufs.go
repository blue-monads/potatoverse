package binds

import (
	"path/filepath"

	"github.com/blue-monads/turnix/backend/engine/executors"
	"github.com/blue-monads/turnix/backend/utils/luaplus"
	lua "github.com/yuin/gopher-lua"
)

// UfsModule provides unified file system bindings for Lua
// Supports three backends: /home (space files), /pkg (package files), /tmp (local fs)
func UfsModule(handle *executors.EHandle) func(L *lua.LState) int {

	return func(L *lua.LState) int {

		// list - List files in a directory
		list := func(L *lua.LState) int {
			path := L.CheckString(1)

			files, err := handle.ListFiles(path)
			if err != nil {
				return pushError(L, err)
			}

			result := L.NewTable()
			for _, file := range files {
				fileTable, err := luaplus.StructToTable(L, file)
				if err != nil {
					return pushError(L, err)
				}

				result.Append(fileTable)
			}

			L.Push(result)
			return 1
		}

		// read - Read file contents
		read := func(L *lua.LState) int {
			path := L.CheckString(1)

			content, err := handle.ReadFile(path)
			if err != nil {
				return pushError(L, err)
			}

			L.Push(lua.LString(string(content)))
			return 1
		}

		// write - Write file contents
		write := func(L *lua.LState) int {
			uid := L.CheckInt64(1)
			path := L.CheckString(2)
			content := []byte(L.CheckString(3))

			err := handle.WriteFile(uid, path, content)
			if err != nil {
				return pushError(L, err)
			}

			L.Push(lua.LTrue)
			return 1
		}

		// remove - Remove a file
		remove := func(L *lua.LState) int {
			uid := L.CheckInt64(1)
			path := L.CheckString(2)

			err := handle.RemoveFile(uid, path)
			if err != nil {
				return pushError(L, err)
			}

			L.Push(lua.LTrue)
			return 1
		}

		// mkdir - Create a directory
		mkdir := func(L *lua.LState) int {
			uid := L.CheckInt64(1)
			path := L.CheckString(2)

			err := handle.Mkdir(uid, path)
			if err != nil {
				return pushError(L, err)
			}

			L.Push(lua.LTrue)
			return 1
		}

		// rmdir - Remove a directory
		rmdir := func(L *lua.LState) int {
			uid := L.CheckInt64(1)
			path := L.CheckString(2)

			err := handle.Rmdir(uid, path)
			if err != nil {
				return pushError(L, err)
			}

			L.Push(lua.LTrue)
			return 1
		}

		// exists - Check if a file or directory exists
		exists := func(L *lua.LState) int {
			path := L.CheckString(1)

			exists, err := handle.Exists(path)
			if err != nil {
				return pushError(L, err)
			}

			L.Push(lua.LBool(exists))
			return 1
		}

		// share - Share a file and get a share link
		share := func(L *lua.LState) int {
			uid := L.CheckInt64(1)
			path := L.CheckString(2)

			shareId, err := handle.ShareFile(uid, path)
			if err != nil {
				return pushError(L, err)
			}

			L.Push(lua.LString(shareId))
			return 1
		}

		// dirname - Get directory name from path
		dirname := func(L *lua.LState) int {
			path := L.CheckString(1)
			dir := filepath.Dir(path)
			L.Push(lua.LString(dir))
			return 1
		}

		// basename - Get base name from path
		basename := func(L *lua.LState) int {
			path := L.CheckString(1)
			base := filepath.Base(path)
			L.Push(lua.LString(base))
			return 1
		}

		// Create and populate the module table
		mod := L.NewTable()

		L.SetFuncs(mod, map[string]lua.LGFunction{
			"list":     list,
			"read":     read,
			"write":    write,
			"remove":   remove,
			"mkdir":    mkdir,
			"rmdir":    rmdir,
			"exists":   exists,
			"share":    share,
			"dirname":  dirname,
			"basename": basename,
		})

		L.Push(mod)
		return 1
	}
}
