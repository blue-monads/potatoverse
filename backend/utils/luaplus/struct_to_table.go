package luaplus

import (
	"fmt"
	"reflect"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// StructToTable converts a Go struct to a Lua table
func StructToTable(l *lua.LState, structValue any) (*lua.LTable, error) {
	if structValue == nil {
		return l.NewTable(), nil
	}

	value := reflect.ValueOf(structValue)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", value.Kind())
	}

	table := l.NewTable()
	valueType := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := valueType.Field(i)
		fieldValue := value.Field(i)

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Get field name (check for json tag first)
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			// Handle json tag (e.g., "name,omitempty")
			fieldName = strings.Split(jsonTag, ",")[0]
			if fieldName == "-" {
				continue // Skip fields marked with "-"
			}
		}

		// Convert field value to Lua value
		luaValue := structFieldToLuaValue(l, fieldValue)
		table.RawSetString(fieldName, luaValue)
	}

	return table, nil
}

// structFieldToLuaValue converts a struct field value to a Lua value
func structFieldToLuaValue(l *lua.LState, fieldValue reflect.Value) lua.LValue {
	if !fieldValue.IsValid() {
		return lua.LNil
	}

	// Handle nil values (only for types that can be nil)
	if fieldValue.Kind() == reflect.Ptr || fieldValue.Kind() == reflect.Slice ||
		fieldValue.Kind() == reflect.Map || fieldValue.Kind() == reflect.Chan ||
		fieldValue.Kind() == reflect.Func || fieldValue.Kind() == reflect.Interface {
		if fieldValue.IsNil() {
			return lua.LNil
		}
	}

	switch fieldValue.Kind() {
	case reflect.String:
		return lua.LString(fieldValue.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(fieldValue.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return lua.LNumber(float64(fieldValue.Uint()))
	case reflect.Float32, reflect.Float64:
		return lua.LNumber(fieldValue.Float())
	case reflect.Bool:
		return lua.LBool(fieldValue.Bool())
	case reflect.Slice:
		// Convert slice to Lua table (array)
		table := l.NewTable()
		for i := 0; i < fieldValue.Len(); i++ {
			itemValue := structFieldToLuaValue(l, fieldValue.Index(i))
			table.RawSetInt(i+1, itemValue) // Lua arrays are 1-indexed
		}
		return table
	case reflect.Map:
		// Convert map to Lua table
		table := l.NewTable()
		for _, key := range fieldValue.MapKeys() {
			value := fieldValue.MapIndex(key)
			keyStr := fmt.Sprintf("%v", key.Interface())
			valueLua := structFieldToLuaValue(l, value)
			table.RawSetString(keyStr, valueLua)
		}
		return table
	case reflect.Struct:
		// Recursively convert nested struct
		nestedTable := l.NewTable()
		nestedType := fieldValue.Type()
		for i := 0; i < fieldValue.NumField(); i++ {
			nestedField := nestedType.Field(i)
			nestedFieldValue := fieldValue.Field(i)

			if !nestedFieldValue.CanInterface() {
				continue
			}

			fieldName := nestedField.Name
			if jsonTag := nestedField.Tag.Get("json"); jsonTag != "" {
				fieldName = strings.Split(jsonTag, ",")[0]
				if fieldName == "-" {
					continue
				}
			}

			fieldLuaValue := structFieldToLuaValue(l, nestedFieldValue)
			nestedTable.RawSetString(fieldName, fieldLuaValue)
		}
		return nestedTable
	case reflect.Ptr:
		// Dereference pointer and convert
		if fieldValue.IsNil() {
			return lua.LNil
		}
		return structFieldToLuaValue(l, fieldValue.Elem())
	default:
		// For other types, convert to string
		return lua.LString(fmt.Sprintf("%v", fieldValue.Interface()))
	}
}
