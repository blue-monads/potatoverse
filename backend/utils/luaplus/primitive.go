package luaplus

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

func LuaTypeToGoType(l *lua.LState, lvalue lua.LValue) any {
	if lvalue == lua.LNil {
		return nil
	}

	switch v := lvalue.(type) {
	case lua.LBool:
		return bool(v)
	case lua.LString:
		return string(v)
	case lua.LNumber:
		// Check if it's an integer
		if v == lua.LNumber(int64(v)) {
			return int64(v)
		}
		return float64(v)
	case *lua.LTable:
		// Check if it's an array (consecutive integer keys starting from 1)
		if isArray(v) {
			return tableToArray(v)
		}
		// Otherwise treat as map
		return tableToMap(v)
	default:
		return nil
	}
}

func GoTypeToLuaType(l *lua.LState, goValue any) lua.LValue {
	if goValue == nil {
		return lua.LNil
	}

	switch v := goValue.(type) {
	case bool:
		return lua.LBool(v)
	case string:
		return lua.LString(v)
	case int:
		return lua.LNumber(v)
	case int64:
		return lua.LNumber(v)
	case float64:
		return lua.LNumber(v)
	case float32:
		return lua.LNumber(v)
	case int8:
		return lua.LNumber(v)
	case int16:
		return lua.LNumber(v)
	case int32:
		return lua.LNumber(v)
	case uint:
		return lua.LNumber(v)
	case uint8:
		return lua.LNumber(v)
	case uint16:
		return lua.LNumber(v)
	case uint32:
		return lua.LNumber(v)
	case uint64:
		return lua.LNumber(v)
	case []byte:
		return lua.LString(v)
	case []any:
		return arrayToTable(l, v)
	case map[string]any:
		return MapToTable(l, v)
	case map[any]any:
		panic("map[any]any not implemented")
	default:
		// For other types, try to convert to string
		return lua.LString(fmt.Sprintf("%v", v))
	}
}

// Helper function to check if a Lua table is an array (consecutive integer keys starting from 1)
func isArray(table *lua.LTable) bool {
	maxKey := 0
	keyCount := 0

	table.ForEach(func(key, value lua.LValue) {
		if num, ok := key.(lua.LNumber); ok {
			if int(num) > 0 {
				keyCount++
				if int(num) > maxKey {
					maxKey = int(num)
				}
			}
		}
	})

	// Array if we have consecutive keys from 1 to maxKey
	return keyCount == maxKey && keyCount > 0
}

// Convert Lua table to Go array
func tableToArray(table *lua.LTable) []any {
	result := make([]any, 0)

	table.ForEach(func(key, value lua.LValue) {
		if num, ok := key.(lua.LNumber); ok {
			idx := int(num) - 1 // Convert to 0-based index
			if idx >= 0 {
				// Ensure slice is large enough
				for len(result) <= idx {
					result = append(result, nil)
				}
				result[idx] = LuaTypeToGoType(nil, value)
			}
		}
	})

	return result
}

// Convert Lua table to Go map
func tableToMap(table *lua.LTable) map[string]any {
	result := make(map[string]any)

	table.ForEach(func(key, value lua.LValue) {
		keyStr := ""
		switch k := key.(type) {
		case lua.LString:
			keyStr = string(k)
		case lua.LNumber:
			keyStr = fmt.Sprintf("%g", float64(k))
		case lua.LBool:
			keyStr = fmt.Sprintf("%t", bool(k))
		default:
			keyStr = fmt.Sprintf("%v", k)
		}
		result[keyStr] = LuaTypeToGoType(nil, value)
	})

	return result
}

// Convert Go array to Lua table
func arrayToTable(l *lua.LState, arr []any) *lua.LTable {
	table := l.NewTable()

	for i, value := range arr {
		table.RawSetInt(i+1, GoTypeToLuaType(l, value)) // Lua arrays are 1-indexed
	}

	return table
}
