package luaplus

import (
	"fmt"
	"reflect"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

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
