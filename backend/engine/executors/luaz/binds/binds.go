package binds

import (
	"encoding/json"
	"strings"

	"github.com/blue-monads/turnix/backend/utils/luaplus"
	lua "github.com/yuin/gopher-lua"
)

type HostHandle interface {
	AddCloser(closer func() error) uint16
	RemoveCloser(id uint16)
}

type LuaLazyData struct {
	table *lua.LTable
	L     *lua.LState
}

func NewLuaLazyData(L *lua.LState, table *lua.LTable) *LuaLazyData {
	return &LuaLazyData{
		L:     L,
		table: table,
	}
}

func (l *LuaLazyData) AsBytes() ([]byte, error) {
	data, err := l.AsMap()
	if err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

func (l *LuaLazyData) AsMap() (map[string]any, error) {
	data := luaplus.TableToMap(l.L, l.table)
	return data, nil
}

func (l *LuaLazyData) AsJson(target any) error {

	err := luaplus.MapToStruct(l.L, l.table, target)
	if err != nil {
		return err
	}

	return nil
}

// getFieldValue navigates the lua table using a dot-separated path and returns the value
func (l *LuaLazyData) getFieldValue(path string) lua.LValue {
	parts := strings.Split(path, ".")
	current := lua.LValue(l.table)

	for _, part := range parts {
		if current == lua.LNil {
			return lua.LNil
		}

		table, ok := current.(*lua.LTable)
		if !ok {
			return lua.LNil
		}

		current = table.RawGetString(part)
	}

	return current
}

func (l *LuaLazyData) GetFieldAsInt(path string) int {
	value := l.getFieldValue(path)
	if value == lua.LNil {
		return 0
	}

	if num, ok := value.(lua.LNumber); ok {
		return int(num)
	}

	return 0
}

func (l *LuaLazyData) GetFieldAsFloat(path string) float64 {
	value := l.getFieldValue(path)
	if value == lua.LNil {
		return 0.0
	}

	if num, ok := value.(lua.LNumber); ok {
		return float64(num)
	}

	return 0.0
}

func (l *LuaLazyData) GetFieldAsString(path string) string {
	value := l.getFieldValue(path)
	if value == lua.LNil {
		return ""
	}

	if str, ok := value.(lua.LString); ok {
		return string(str)
	}

	return ""
}

func (l *LuaLazyData) GetFieldAsBool(path string) bool {
	value := l.getFieldValue(path)
	if value == lua.LNil {
		return false
	}

	if b, ok := value.(lua.LBool); ok {
		return bool(b)
	}

	return false
}

func pushError(L *lua.LState, err error) int {
	return luaplus.PushError(L, err)
}
