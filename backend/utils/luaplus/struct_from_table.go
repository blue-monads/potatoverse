package luaplus

import (
	"fmt"
	"reflect"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// MapToStruct maps a Lua table into a Go struct via reflection.
// Field matching priority: `lua` tag > `json` tag > exact field name > snake_case of field name.
func MapToStruct(l *lua.LState, lvalue lua.LValue, target any) error {
	if lvalue == lua.LNil {
		return fmt.Errorf("cannot convert nil to struct")
	}

	table, ok := lvalue.(*lua.LTable)
	if !ok {
		return fmt.Errorf("expected table, got %T", lvalue)
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer to a struct")
	}

	targetValue = targetValue.Elem()
	if targetValue.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct, got pointer to %s", targetValue.Kind())
	}

	return mapTableToStruct(table, targetValue)
}

// mapTableToStruct fills a struct reflect.Value from a Lua table.
func mapTableToStruct(table *lua.LTable, structVal reflect.Value) error {
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		field := structType.Field(i)
		fieldVal := structVal.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		fieldName := luaFieldName(field)
		if fieldName == "-" {
			continue
		}

		// Try the tagged/exact name first, then snake_case fallback.
		lv := table.RawGetString(fieldName)
		if lv == lua.LNil {
			snake := toSnakeCase(fieldName)
			if snake != fieldName {
				lv = table.RawGetString(snake)
			}
			if lv == lua.LNil {
				continue
			}
		}
		fmt.Printf(" -> HIT value=%s\n", lv.String())

		if err := setField(fieldVal, lv); err != nil {
			return fmt.Errorf("field %q: %w", field.Name, err)
		}
	}

	return nil
}

// luaFieldName resolves the lookup key for a struct field.
// Priority: lua tag > json tag > field name. Returns "-" for explicitly skipped fields.
func luaFieldName(f reflect.StructField) string {
	if tag := f.Tag.Get("lua"); tag != "" {
		return strings.SplitN(tag, ",", 2)[0]
	}
	if tag := f.Tag.Get("json"); tag != "" {
		return strings.SplitN(tag, ",", 2)[0]
	}
	return f.Name
}

// toSnakeCase converts PascalCase/camelCase to snake_case.
// "LeftTable" -> "left_table", "URLPath" -> "url_path"
func toSnakeCase(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		upper := r >= 'A' && r <= 'Z'
		if !upper {
			b.WriteRune(r)
			continue
		}
		if i > 0 {
			prevLower := runes[i-1] >= 'a' && runes[i-1] <= 'z'
			nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
			if prevLower || nextLower {
				b.WriteRune('_')
			}
		}
		b.WriteRune(r + 32)
	}
	return b.String()
}

// setField converts a Lua value and assigns it to a reflect.Value.
func setField(dst reflect.Value, lv lua.LValue) error {
	// Unwrap pointer - allocate if nil
	if dst.Kind() == reflect.Ptr {
		if lv == lua.LNil {
			dst.Set(reflect.Zero(dst.Type()))
			return nil
		}
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return setField(dst.Elem(), lv)
	}

	switch dst.Kind() {
	case reflect.String:
		return setString(dst, lv)
	case reflect.Bool:
		return setBool(dst, lv)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setInt(dst, lv)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return setUint(dst, lv)
	case reflect.Float32, reflect.Float64:
		return setFloat(dst, lv)
	case reflect.Slice:
		return setSlice(dst, lv)
	case reflect.Map:
		return setMap(dst, lv)
	case reflect.Struct:
		return setStruct(dst, lv)
	case reflect.Interface:
		dst.Set(reflect.ValueOf(luaToAny(lv)))
		return nil
	default:
		return fmt.Errorf("unsupported kind %s", dst.Kind())
	}
}

// -- scalar setters -----------------------------------------------------------

func setString(dst reflect.Value, lv lua.LValue) error {
	switch v := lv.(type) {
	case lua.LString:
		dst.SetString(string(v))
	case lua.LNumber:
		dst.SetString(fmt.Sprintf("%g", float64(v)))
	case lua.LBool:
		dst.SetString(fmt.Sprintf("%t", bool(v)))
	default:
		dst.SetString(lv.String())
	}
	return nil
}

func setBool(dst reflect.Value, lv lua.LValue) error {
	b, ok := lv.(lua.LBool)
	if !ok {
		return fmt.Errorf("expected bool, got %s", lv.Type())
	}
	dst.SetBool(bool(b))
	return nil
}

func setInt(dst reflect.Value, lv lua.LValue) error {
	n, ok := lv.(lua.LNumber)
	if !ok {
		return fmt.Errorf("expected number, got %s", lv.Type())
	}
	dst.SetInt(int64(n))
	return nil
}

func setUint(dst reflect.Value, lv lua.LValue) error {
	n, ok := lv.(lua.LNumber)
	if !ok {
		return fmt.Errorf("expected number, got %s", lv.Type())
	}
	if float64(n) < 0 {
		return fmt.Errorf("cannot assign negative number %g to unsigned field", float64(n))
	}
	dst.SetUint(uint64(n))
	return nil
}

func setFloat(dst reflect.Value, lv lua.LValue) error {
	n, ok := lv.(lua.LNumber)
	if !ok {
		return fmt.Errorf("expected number, got %s", lv.Type())
	}
	dst.SetFloat(float64(n))
	return nil
}

// -- composite setters --------------------------------------------------------

// setSlice handles Lua array tables -> Go slices.
func setSlice(dst reflect.Value, lv lua.LValue) error {
	table, ok := lv.(*lua.LTable)
	if !ok {
		return fmt.Errorf("expected table for slice, got %s", lv.Type())
	}

	fmt.Printf("[setSlice] elem type=%s table.Len()=%d\n", dst.Type().Elem(), table.Len())

	var items []lua.LValue
	arrayLen := table.Len()

	if arrayLen > 0 {
		for i := 1; i <= arrayLen; i++ {
			items = append(items, table.RawGetInt(i))
		}
	} else {
		table.ForEach(func(_, v lua.LValue) {
			items = append(items, v)
		})
	}

	fmt.Printf("[setSlice] collected %d items\n", len(items))

	elemType := dst.Type().Elem()
	slice := reflect.MakeSlice(dst.Type(), len(items), len(items))

	for i, item := range items {
		fmt.Printf("[setSlice] item[%d] type=%s value=%s\n", i, item.Type(), item.String())
		elem := slice.Index(i)
		if elemType.Kind() == reflect.Ptr {
			ptr := reflect.New(elemType.Elem())
			if err := setField(ptr.Elem(), item); err != nil {
				return fmt.Errorf("slice[%d]: %w", i, err)
			}
			elem.Set(ptr)
			continue
		}
		if err := setField(elem, item); err != nil {
			return fmt.Errorf("slice[%d]: %w", i, err)
		}
	}

	dst.Set(slice)
	return nil
}

// setMap handles Lua tables as map[K]V.
// Key types supported: string, all int/uint/float kinds, and interface{}.
func setMap(dst reflect.Value, lv lua.LValue) error {
	table, ok := lv.(*lua.LTable)
	if !ok {
		return fmt.Errorf("expected table for map, got %s", lv.Type())
	}

	mapType := dst.Type()
	keyKind := mapType.Key().Kind()

	newMap := reflect.MakeMap(mapType)
	var iterErr error

	table.ForEach(func(k, v lua.LValue) {
		if iterErr != nil {
			return
		}

		keyVal, err := convertLuaKey(k, mapType.Key())
		if err != nil {
			iterErr = fmt.Errorf("map key (kind %s): %w", keyKind, err)
			return
		}

		elem := reflect.New(mapType.Elem()).Elem()
		if err := setField(elem, v); err != nil {
			iterErr = fmt.Errorf("map key %v: %w", k, err)
			return
		}
		newMap.SetMapIndex(keyVal, elem)
	})

	if iterErr != nil {
		return iterErr
	}

	dst.Set(newMap)
	return nil
}

// convertLuaKey converts a Lua key value to a reflect.Value of the target key type.
func convertLuaKey(lk lua.LValue, keyType reflect.Type) (reflect.Value, error) {
	keyKind := keyType.Kind()

	// interface{} key - store the natural Go equivalent
	if keyKind == reflect.Interface {
		return reflect.ValueOf(luaToAny(lk)), nil
	}

	switch keyKind {
	case reflect.String:
		switch v := lk.(type) {
		case lua.LString:
			return reflect.ValueOf(string(v)).Convert(keyType), nil
		case lua.LNumber:
			return reflect.ValueOf(fmt.Sprintf("%g", float64(v))).Convert(keyType), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot use %s as string key", lk.Type())
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, ok := lk.(lua.LNumber)
		if !ok {
			return reflect.Value{}, fmt.Errorf("cannot use %s as integer key", lk.Type())
		}
		kv := reflect.New(keyType).Elem()
		kv.SetInt(int64(n))
		return kv, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, ok := lk.(lua.LNumber)
		if !ok {
			return reflect.Value{}, fmt.Errorf("cannot use %s as unsigned integer key", lk.Type())
		}
		if float64(n) < 0 {
			return reflect.Value{}, fmt.Errorf("negative number %g cannot be used as unsigned key", float64(n))
		}
		kv := reflect.New(keyType).Elem()
		kv.SetUint(uint64(n))
		return kv, nil

	case reflect.Float32, reflect.Float64:
		n, ok := lk.(lua.LNumber)
		if !ok {
			return reflect.Value{}, fmt.Errorf("cannot use %s as float key", lk.Type())
		}
		kv := reflect.New(keyType).Elem()
		kv.SetFloat(float64(n))
		return kv, nil

	default:
		return reflect.Value{}, fmt.Errorf("unsupported map key kind: %s", keyKind)
	}
}

// setStruct handles Lua tables as nested structs.
func setStruct(dst reflect.Value, lv lua.LValue) error {
	table, ok := lv.(*lua.LTable)
	if !ok {
		return fmt.Errorf("expected table for struct, got %s", lv.Type())
	}
	return mapTableToStruct(table, dst)
}

// -- helpers ------------------------------------------------------------------

// luaToAny converts a Lua value to a plain Go any, used for interface{} fields.
func luaToAny(lv lua.LValue) any {
	switch v := lv.(type) {
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LTable:
		if v.Len() > 0 {
			out := make([]any, v.Len())
			for i := 1; i <= v.Len(); i++ {
				out[i-1] = luaToAny(v.RawGetInt(i))
			}
			return out
		}
		out := make(map[string]any)
		v.ForEach(func(k, val lua.LValue) {
			if key, ok := k.(lua.LString); ok {
				out[string(key)] = luaToAny(val)
			}
		})
		return out
	default:
		return nil
	}
}
