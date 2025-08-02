package blua

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

func (b *Blua) exists(L *lua.LState) int {
	path := L.CheckString(1)
	_, err := b.handle.baseFolder.Stat(path)

	if os.IsNotExist(err) {
		L.Push(lua.LBool(false))
	} else {
		L.Push(lua.LBool(true))
	}

	return 1
}

func (b *Blua) read(L *lua.LState) int {
	path := L.CheckString(1)

	file, err := b.handle.baseFolder.Open(path)
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

func (b *Blua) write(L *lua.LState) int {
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

	file, err := b.handle.baseFolder.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
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

func (b *Blua) mkdir(L *lua.LState) int {
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
	err := b.handle.baseFolder.Mkdir(dir, mode)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	return 1
}

func (b *Blua) remove(L *lua.LState) int {
	path := L.CheckString(1)

	err := b.handle.baseFolder.Remove(path)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	return 1
}

func (b *Blua) dirname(L *lua.LState) int {
	filePath := L.CheckString(1)
	dirPath := filepath.Dir(filePath)
	L.Push(lua.LString(dirPath))
	return 1
}

func (b *Blua) basename(L *lua.LState) int {
	filePath := L.CheckString(1)
	baseName := filepath.Base(filePath)
	L.Push(lua.LString(baseName))
	return 1
}

func (b *Blua) fnRealpath(L *lua.LState) int {
	// Since we're operating in a scoped filesystem,
	// realpath doesn't make much sense in this context
	// We'll return the path relative to the base folder
	path := L.CheckString(1)

	// Check if file exists
	_, err := b.handle.baseFolder.Stat(path)
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

func (b *Blua) glob(L *lua.LState) int {
	// Note: Root doesn't have a direct glob method
	// This would need a custom implementation using the Open and ReadDir methods
	// For simplicity, returning an error
	L.Push(lua.LNil)
	L.Push(lua.LString("glob operation not supported in scoped filesystem"))
	return 2
}
