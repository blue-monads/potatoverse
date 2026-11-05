package luaplus

import (
	lua "github.com/yuin/gopher-lua"
)

func TableToArray(L *lua.LState, table *lua.LTable) []any {
	result := make([]any, 0)

	table.ForEach(func(_ lua.LValue, value lua.LValue) {
		result = append(result, LuaTypeToGoType(L, value))
	})

	return result
}

func TableToMap(L *lua.LState, table *lua.LTable) map[string]any {
	result := make(map[string]any)

	table.ForEach(func(key, value lua.LValue) {
		result[key.String()] = LuaToAny(L, value)
	})

	return result
}

func LuaToAny(L *lua.LState, value lua.LValue) any {
	switch value.Type() {
	case lua.LTString:
		return value.String()
	case lua.LTNumber:
		return float64(value.(lua.LNumber))
	case lua.LTBool:
		return bool(value.(lua.LBool))
	case lua.LTTable:
		tbl := value.(*lua.LTable)
		if isTableEmpty(tbl) {
			return nil
		}
		if isArray(tbl) {
			return tableToArray(tbl)
		} else {
			return TableToMap(L, tbl)
		}
	case lua.LTNil:
		return nil

	default:
		return value.String()
	}

}

func TableToMapAny(L *lua.LState, table *lua.LTable) map[any]any {
	result := make(map[any]any)

	table.ForEach(func(key, value lua.LValue) {
		result[key.String()] = LuaToAny(L, value)
	})

	return result
}

func MapToTable(L *lua.LState, m map[string]any) *lua.LTable {
	table := L.NewTable()

	for k, v := range m {
		L.SetField(table, k, GoTypeToLuaType(L, v))
	}

	return table
}
