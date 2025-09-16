package binds

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

func MapToTable(L *lua.LState, m map[string]any) *lua.LTable {
	table := L.NewTable()

	for k, v := range m {
		switch vt := v.(type) {
		case string:
			L.SetField(table, k, lua.LString(vt))
		case float64:
			L.SetField(table, k, lua.LNumber(vt))
		case int:
			L.SetField(table, k, lua.LNumber(vt))
		case bool:
			L.SetField(table, k, lua.LBool(vt))
		case map[string]any:
			L.SetField(table, k, MapToTable(L, vt))
		case []any:
			arrayTable := L.NewTable()
			for i, item := range vt {
				switch it := item.(type) {
				case string:
					L.RawSetInt(arrayTable, i+1, lua.LString(it))
				case float64:
					L.RawSetInt(arrayTable, i+1, lua.LNumber(it))
				case int:
					L.RawSetInt(arrayTable, i+1, lua.LNumber(it))
				case bool:
					L.RawSetInt(arrayTable, i+1, lua.LBool(it))
				case map[string]any:
					L.RawSetInt(arrayTable, i+1, MapToTable(L, it))
				default:
					L.RawSetInt(arrayTable, i+1, lua.LString(fmt.Sprintf("%v", it)))
				}
			}
			L.SetField(table, k, arrayTable)
		default:
			L.SetField(table, k, lua.LString(fmt.Sprintf("%v", vt)))
		}
	}

	return table
}

// Helper function to convert Lua tables to Go maps
func TableToMap(L *lua.LState, table *lua.LTable) map[string]any {
	result := make(map[string]any)

	table.ForEach(func(key, value lua.LValue) {
		switch value.Type() {
		case lua.LTString:
			result[key.String()] = value.String()
		case lua.LTNumber:
			result[key.String()] = float64(value.(lua.LNumber))
		case lua.LTBool:
			result[key.String()] = bool(value.(lua.LBool))
		case lua.LTTable:
			result[key.String()] = TableToMap(L, value.(*lua.LTable))
		default:
			result[key.String()] = value.String()
		}
	})

	return result
}
