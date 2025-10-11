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

func TableToMapAny(L *lua.LState, table *lua.LTable) map[any]any {
	result := make(map[any]any)

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

func toStructFromTableInner(l *lua.LState, tb *lua.LTable, v reflect.Value) error {
	// If the value is a pointer, dereference it.
	// This is the key change to fix the panic.
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Now check if the dereferenced value is a struct.
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct, but got %s", v.Kind())
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		var inline bool
		name := field.Name
		if tag := field.Tag.Get("luautil"); tag != "" {
			tagParts := strings.Split(tag, ",")
			if tagParts[0] == "-" {
				continue // Skip this field.
			} else if tagParts[0] != "" {
				name = tagParts[0] // Use the name from the tag.
			}
			if len(tagParts) > 1 && tagParts[1] == "inline" {
				inline = true // Handle inlining.
			}
		}

		goField := v.Field(i)
		if !goField.CanSet() {
			continue
		}

		if inline {
			// If it's an inline struct, recurse into it.
			if goField.Kind() == reflect.Struct {
				if err := toStructFromTableInner(l, tb, goField); err != nil {
					return err
				}
			}
			continue
		}

		// Look for the corresponding value in the Lua table.
		luaValue := tb.RawGetString(name)
		if luaValue == lua.LNil {
			continue // No matching key.
		}

		// Convert the Lua value to the Go field's type.
		if err := setGoFieldValue(goField, luaValue); err != nil {
			return fmt.Errorf("failed to set field '%s': %w", name, err)
		}
	}
	return nil
}

// setGoFieldValue handles the type conversion and assignment.
func setGoFieldValue(goField reflect.Value, luaValue lua.LValue) error {
	switch goField.Kind() {
	case reflect.String:
		if str, ok := luaValue.(lua.LString); ok {
			goField.SetString(string(str))
			return nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if num, ok := luaValue.(lua.LNumber); ok {
			goField.SetInt(int64(num))
			return nil
		}
	case reflect.Float32, reflect.Float64:
		if num, ok := luaValue.(lua.LNumber); ok {
			goField.SetFloat(float64(num))
			return nil
		}
	case reflect.Bool:
		if b, ok := luaValue.(lua.LBool); ok {
			goField.SetBool(bool(b))
			return nil
		}
	case reflect.Struct:
		// Handle nested structs.
		if nestedTb, ok := luaValue.(*lua.LTable); ok {
			return toStructFromTableInner(nil, nestedTb, goField)
		}
	}
	return fmt.Errorf("unsupported type conversion from %T to %s", luaValue, goField.Kind())
}

func ToTableFromStruct(l *lua.LState, v reflect.Value) lua.LValue {
	tb := l.NewTable()
	return toTableFromStructInner(l, tb, v)
}

func toTableFromStructInner(l *lua.LState, tb *lua.LTable, v reflect.Value) lua.LValue {
	t := v.Type()
	for j := 0; j < v.NumField(); j++ {
		field := t.Field(j)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		var inline bool
		name := field.Name
		if tag := field.Tag.Get("luautil"); tag != "" {
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
