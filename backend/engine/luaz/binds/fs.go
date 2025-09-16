package binds

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	lua "github.com/yuin/gopher-lua"
)

// Helper function to convert octal to decimal
func oct2decimal(oct int) (uint64, error) {
	return strconv.ParseUint(fmt.Sprintf("%d", oct), 8, 32)
}

func FsModule(root *os.Root) func(L *lua.LState) int {

	return func(L *lua.LState) int {

		exists := func(L *lua.LState) int {
			path := L.CheckString(1)
			_, err := root.Stat(path)

			if os.IsNotExist(err) {
				L.Push(lua.LBool(false))
			} else {
				L.Push(lua.LBool(true))
			}

			return 1
		}

		read := func(L *lua.LState) int {
			path := L.CheckString(1)

			file, err := root.Open(path)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			defer file.Close()

			content, err := io.ReadAll(file)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(lua.LString(string(content)))
			return 1
		}

		write := func(L *lua.LState) int {
			path := L.CheckString(1)
			content := []byte(L.CheckString(2))
			var mode os.FileMode = 0666

			top := L.GetTop()
			if top == 3 {
				m, err := oct2decimal(L.CheckInt(3))
				if err != nil {
					L.Push(lua.LNil)
					L.Push(lua.LString(err.Error()))
					return 2
				}
				mode = os.FileMode(m)
			}

			file, err := root.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			defer file.Close()

			_, err = file.Write(content)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(lua.LTrue)
			return 1
		}

		mkdir := func(L *lua.LState) int {
			dir := L.CheckString(1)
			var mode os.FileMode = 0777

			top := L.GetTop()
			if top >= 2 {
				m, err := oct2decimal(L.CheckInt(2))
				if err != nil {
					L.Push(lua.LNil)
					L.Push(lua.LString(err.Error()))
					return 2
				}
				mode = os.FileMode(m)
			}

			// Note: Root doesn't have MkdirAll, so recursive creation is not supported
			// Unless we implement it manually
			err := root.Mkdir(dir, mode)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(lua.LTrue)
			return 1
		}

		remove := func(L *lua.LState) int {
			path := L.CheckString(1)

			err := root.Remove(path)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			L.Push(lua.LTrue)
			return 1
		}

		dirname := func(L *lua.LState) int {
			filePath := L.CheckString(1)
			dirPath := filepath.Dir(filePath)
			L.Push(lua.LString(dirPath))
			return 1
		}

		basename := func(L *lua.LState) int {
			filePath := L.CheckString(1)
			baseName := filepath.Base(filePath)
			L.Push(lua.LString(baseName))
			return 1
		}

		fnRealpath := func(L *lua.LState) int {
			// Since we're operating in a scoped filesystem,
			// realpath doesn't make much sense in this context
			// We'll return the path relative to the base folder
			path := L.CheckString(1)

			// Check if file exists
			_, err := root.Stat(path)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			// Return the cleaned path
			cleanPath := filepath.Clean(path)
			L.Push(lua.LString(cleanPath))
			return 1
		}

		mod := L.NewTable()

		L.SetFuncs(mod, map[string]lua.LGFunction{
			"exists":     exists,
			"read":       read,
			"write":      write,
			"fnRealpath": fnRealpath,
			"basename":   basename,
			"dirname":    dirname,
			"remove":     remove,
			"mkdir":      mkdir,
		})

		return 0
	}

}
