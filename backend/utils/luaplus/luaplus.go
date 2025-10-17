package luaplus

import (
	"fmt"
	"reflect"
	"strings"

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
	case []any:
		return arrayToTable(l, v)
	case map[string]any:
		return mapToTable(l, v)
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

// Convert Go map to Lua table
func mapToTable(l *lua.LState, m map[string]any) *lua.LTable {
	table := l.NewTable()

	for key, value := range m {
		table.RawSetString(key, GoTypeToLuaType(l, value))
	}

	return table
}

func MapToStruct(l *lua.LState, lvalue lua.LValue, target any) error {
	if lvalue == lua.LNil {
		return fmt.Errorf("cannot convert nil to struct")
	}

	table, ok := lvalue.(*lua.LTable)
	if !ok {
		return fmt.Errorf("expected table, got %T", lvalue)
	}

	// Get the target's reflection value
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer to struct")
	}

	// Dereference the pointer
	targetValue = targetValue.Elem()
	if targetValue.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	// Convert Lua table to Go map first
	luaMap := tableToMap(table)

	// Use reflection to set struct fields
	targetType := targetValue.Type()
	for i := 0; i < targetValue.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := targetValue.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get field name (check for json tag first)
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			// Handle json tag (e.g., "name,omitempty")
			fieldName = strings.Split(jsonTag, ",")[0]
		}

		// Get value from Lua table
		luaValue, exists := luaMap[fieldName]
		if !exists {
			continue // Skip if field not present in Lua table
		}

		// Convert and set the field value
		if err := setStructField(fieldValue, luaValue); err != nil {
			return fmt.Errorf("failed to set field %s: %w", fieldName, err)
		}
	}

	return nil
}

// setStructField sets a struct field value from a Go value
func setStructField(fieldValue reflect.Value, goValue any) error {
	if !fieldValue.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	// Handle nil values
	if goValue == nil {
		fieldValue.Set(reflect.Zero(fieldValue.Type()))
		return nil
	}

	// Get the type of the field
	fieldType := fieldValue.Type()
	valueType := reflect.TypeOf(goValue)

	// Direct type match
	if valueType.AssignableTo(fieldType) {
		fieldValue.Set(reflect.ValueOf(goValue))
		return nil
	}

	// Handle type conversions
	switch fieldType.Kind() {
	case reflect.String:
		if str, ok := goValue.(string); ok {
			fieldValue.SetString(str)
		} else {
			fieldValue.SetString(fmt.Sprintf("%v", goValue))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if num, ok := goValue.(float64); ok {
			fieldValue.SetInt(int64(num))
		} else if num, ok := goValue.(int64); ok {
			fieldValue.SetInt(num)
		} else if num, ok := goValue.(int); ok {
			fieldValue.SetInt(int64(num))
		} else {
			return fmt.Errorf("cannot convert %T to %s", goValue, fieldType.Kind())
		}
	case reflect.Float32, reflect.Float64:
		if num, ok := goValue.(float64); ok {
			fieldValue.SetFloat(num)
		} else if num, ok := goValue.(int64); ok {
			fieldValue.SetFloat(float64(num))
		} else if num, ok := goValue.(int); ok {
			fieldValue.SetFloat(float64(num))
		} else {
			return fmt.Errorf("cannot convert %T to %s", goValue, fieldType.Kind())
		}
	case reflect.Bool:
		if b, ok := goValue.(bool); ok {
			fieldValue.SetBool(b)
		} else {
			return fmt.Errorf("cannot convert %T to bool", goValue)
		}
	case reflect.Slice:
		if slice, ok := goValue.([]any); ok {
			// Create a new slice of the correct type
			newSlice := reflect.MakeSlice(fieldType, len(slice), len(slice))
			for i, item := range slice {
				if i < newSlice.Len() {
					// Convert each item
					itemValue := reflect.ValueOf(item)
					if itemValue.Type().AssignableTo(fieldType.Elem()) {
						newSlice.Index(i).Set(itemValue)
					} else {
						// Try to convert the item
						if err := setStructField(newSlice.Index(i), item); err != nil {
							return fmt.Errorf("cannot convert slice item %d: %w", i, err)
						}
					}
				}
			}
			fieldValue.Set(newSlice)
		} else {
			return fmt.Errorf("cannot convert %T to slice", goValue)
		}
	case reflect.Map:
		if m, ok := goValue.(map[string]any); ok {
			// Create a new map of the correct type
			newMap := reflect.MakeMap(fieldType)
			for key, value := range m {
				keyValue := reflect.ValueOf(key)
				valueValue := reflect.ValueOf(value)

				// Convert value if needed
				if !valueValue.Type().AssignableTo(fieldType.Elem()) {
					// Create a new value of the correct type
					newValue := reflect.New(fieldType.Elem()).Elem()
					if err := setStructField(newValue, value); err != nil {
						return fmt.Errorf("cannot convert map value for key %s: %w", key, err)
					}
					valueValue = newValue
				}

				newMap.SetMapIndex(keyValue, valueValue)
			}
			fieldValue.Set(newMap)
		} else {
			return fmt.Errorf("cannot convert %T to map", goValue)
		}
	case reflect.Struct:
		if m, ok := goValue.(map[string]any); ok {
			// Recursively convert nested struct
			for key, value := range m {
				// Find the field by name or json tag
				fieldFound := false
				for i := 0; i < fieldType.NumField(); i++ {
					field := fieldType.Field(i)
					fieldName := field.Name
					if jsonTag := field.Tag.Get("json"); jsonTag != "" {
						fieldName = strings.Split(jsonTag, ",")[0]
					}

					if fieldName == key {
						nestedField := fieldValue.Field(i)
						if nestedField.CanSet() {
							if err := setStructField(nestedField, value); err != nil {
								return fmt.Errorf("cannot set nested field %s: %w", key, err)
							}
							fieldFound = true
							break
						}
					}
				}
				if !fieldFound {
					// Skip unknown fields
					continue
				}
			}
		} else {
			return fmt.Errorf("cannot convert %T to struct", goValue)
		}
	default:
		return fmt.Errorf("unsupported field type: %s", fieldType.Kind())
	}

	return nil
}

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
