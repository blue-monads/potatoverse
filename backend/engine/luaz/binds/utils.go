package binds

import (
	"fmt"
	"reflect"
	"strings"

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

func ToTableFromStruct(l *lua.LState, v reflect.Value) lua.LValue {
	tb := l.NewTable()
	return toTableFromStructInner(l, tb, v)
}

func toTableFromStructInner(l *lua.LState, tb *lua.LTable, v reflect.Value) lua.LValue {
	t := v.Type()
	for j := 0; j < v.NumField(); j++ {
		var inline bool
		name := t.Field(j).Name
		if tag := t.Field(j).Tag.Get("luautil"); tag != "" {
			tagParts := strings.Split(tag, ",")
			if tagParts[0] == "-" {
				continue
			} else if tagParts[0] != "" {
				name = tagParts[0]
			}
			if len(tagParts) > 1 && tagParts[1] == "inline" {
				inline = true
			}
		}
		if inline {
			toTableFromStructInner(l, tb, v.Field(j))
		} else {
			tb.RawSetString(name, ToArbitraryValue(l, v.Field(j).Interface()))
		}
	}
	return tb
}

func ToTableFromMap(l *lua.LState, v reflect.Value) lua.LValue {
	tb := &lua.LTable{}
	for _, k := range v.MapKeys() {
		tb.RawSet(ToArbitraryValue(l, k.Interface()),
			ToArbitraryValue(l, v.MapIndex(k).Interface()))
	}
	return tb
}

func ToTableFromSlice(l *lua.LState, v reflect.Value) lua.LValue {
	tb := &lua.LTable{}
	for j := 0; j < v.Len(); j++ {
		tb.RawSet(ToArbitraryValue(l, j+1), // because lua is 1-indexed
			ToArbitraryValue(l, v.Index(j).Interface()))
	}
	return tb
}

// ToArbitraryValue converts Go values to Lua values
func ToArbitraryValue(l *lua.LState, i any) lua.LValue {
	if i == nil {
		return lua.LNil
	}

	switch ii := i.(type) {
	case bool:
		return lua.LBool(ii)
	case int:
		return lua.LNumber(ii)
	case int8:
		return lua.LNumber(ii)
	case int16:
		return lua.LNumber(ii)
	case int32:
		return lua.LNumber(ii)
	case int64:
		return lua.LNumber(ii)
	case uint:
		return lua.LNumber(ii)
	case uint8:
		return lua.LNumber(ii)
	case uint16:
		return lua.LNumber(ii)
	case uint32:
		return lua.LNumber(ii)
	case uint64:
		return lua.LNumber(ii)
	case float64:
		return lua.LNumber(ii)
	case float32:
		return lua.LNumber(ii)
	case string:
		return lua.LString(ii)
	case []byte:
		return lua.LString(ii)
	default:
		v := reflect.ValueOf(i)
		switch v.Kind() {
		case reflect.Ptr:
			return ToArbitraryValue(l, v.Elem().Interface())
		case reflect.Struct:
			return ToTableFromStruct(l, v)
		case reflect.Map:
			return ToTableFromMap(l, v)
		case reflect.Slice:
			return ToTableFromSlice(l, v)
		default:
			// Handle sql.RawBytes specifically, which is commonly returned by database/sql
			return lua.LString(fmt.Sprintf("%v", i))
		}
	}
}
