package stateless

import (
	"fmt"
	"log"

	lua "github.com/yuin/gopher-lua"
)

/*

lua type to go type
	- int <-> int64
	- float <-> float64
	- string <-> string
	- bool <-> bool
	- table <-> map[string]any
	- array <-> []any


*/

func RunStateLessLua() {
	err := runStateLessLua()
	if err != nil {
		log.Fatalf("Failed to run state less Lua: %v", err)
	}
}

func runStateLessLua() error {
	const testcode = `
-- Test basic types
local test_nil = nil
local test_bool = true
local test_string = "hello world"
local test_number_int = 42
local test_number_float = 3.14

-- Test array
local test_array = {1, 2, 3, "four", true}

-- Test table/map
local test_table = {
    name = "John",
    age = 30,
    active = true,
    score = 95.5
}

-- Test nested structures
local test_nested = {
    user = {
        id = 1,
        name = "Alice",
        tags = {"admin", "user", "premium"}
    },
    settings = {
        theme = "dark",
        notifications = true
    }
}

-- Test mixed array/table
local test_mixed = {
    "first", "second",  -- array part
    name = "mixed",     -- table part
    count = 2
}

-- Store test results in global variables for Go to access
_G.test_nil = test_nil
_G.test_bool = test_bool
_G.test_string = test_string
_G.test_number_int = test_number_int
_G.test_number_float = test_number_float
_G.test_array = test_array
_G.test_table = test_table
_G.test_nested = test_nested
_G.test_mixed = test_mixed
`

	L := lua.NewState()
	defer L.Close()

	err := L.DoString(testcode)
	if err != nil {
		return err
	}

	// Test basic type conversions
	fmt.Println("=== Testing Basic Type Conversions ===")

	// Test nil
	testNil := L.GetGlobal("test_nil")
	goNil := LuaTypeToGoType(L, testNil)
	fmt.Printf("Lua nil -> Go: %v (type: %T)\n", goNil, goNil)

	// Test bool
	testBool := L.GetGlobal("test_bool")
	goBool := LuaTypeToGoType(L, testBool)
	fmt.Printf("Lua bool -> Go: %v (type: %T)\n", goBool, goBool)

	// Test string
	testString := L.GetGlobal("test_string")
	goString := LuaTypeToGoType(L, testString)
	fmt.Printf("Lua string -> Go: %v (type: %T)\n", goString, goString)

	// Test number (int)
	testNumberInt := L.GetGlobal("test_number_int")
	goNumberInt := LuaTypeToGoType(L, testNumberInt)
	fmt.Printf("Lua number (int) -> Go: %v (type: %T)\n", goNumberInt, goNumberInt)

	// Test number (float)
	testNumberFloat := L.GetGlobal("test_number_float")
	goNumberFloat := LuaTypeToGoType(L, testNumberFloat)
	fmt.Printf("Lua number (float) -> Go: %v (type: %T)\n", goNumberFloat, goNumberFloat)

	// Test array conversion
	fmt.Println("\n=== Testing Array Conversion ===")
	testArray := L.GetGlobal("test_array")
	goArray := LuaTypeToGoType(L, testArray)
	fmt.Printf("Lua array -> Go: %v (type: %T)\n", goArray, goArray)

	// Test table conversion
	fmt.Println("\n=== Testing Table Conversion ===")
	testTable := L.GetGlobal("test_table")
	goTable := LuaTypeToGoType(L, testTable)
	fmt.Printf("Lua table -> Go: %v (type: %T)\n", goTable, goTable)

	// Test nested structure conversion
	fmt.Println("\n=== Testing Nested Structure Conversion ===")
	testNested := L.GetGlobal("test_nested")
	goNested := LuaTypeToGoType(L, testNested)
	fmt.Printf("Lua nested -> Go: %v (type: %T)\n", goNested, goNested)

	// Test mixed array/table
	fmt.Println("\n=== Testing Mixed Array/Table Conversion ===")
	testMixed := L.GetGlobal("test_mixed")
	goMixed := LuaTypeToGoType(L, testMixed)
	fmt.Printf("Lua mixed -> Go: %v (type: %T)\n", goMixed, goMixed)

	// Test Go to Lua conversions
	fmt.Println("\n=== Testing Go to Lua Conversions ===")

	// Test Go array to Lua
	goTestArray := []any{1, 2, 3, "four", true}
	luaArray := GoTypeToLuaType(L, goTestArray)
	fmt.Printf("Go array -> Lua: %v\n", luaArray)

	// Test Go map to Lua
	goTestMap := map[string]any{
		"name":   "Bob",
		"age":    25,
		"active": false,
		"score":  87.3,
	}
	luaMap := GoTypeToLuaType(L, goTestMap)
	fmt.Printf("Go map -> Lua: %v\n", luaMap)

	// Test nested Go structure to Lua
	goTestNested := map[string]any{
		"user": map[string]any{
			"id":   1,
			"name": "Charlie",
			"tags": []any{"user", "beta"},
		},
		"settings": map[string]any{
			"theme":         "light",
			"notifications": false,
		},
	}
	luaNested := GoTypeToLuaType(L, goTestNested)
	fmt.Printf("Go nested -> Lua: %v\n", luaNested)

	// Test round-trip conversions
	fmt.Println("\n=== Testing Round-trip Conversions ===")

	// Round-trip: Lua -> Go -> Lua
	originalLuaValue := L.GetGlobal("test_table")
	goValue := LuaTypeToGoType(L, originalLuaValue)
	backToLua := GoTypeToLuaType(L, goValue)
	fmt.Printf("Round-trip (Lua->Go->Lua): original=%v, back=%v\n", originalLuaValue, backToLua)

	// Round-trip: Go -> Lua -> Go
	originalGoValue := []any{"test", 42, true}
	luaValue := GoTypeToLuaType(L, originalGoValue)
	backToGo := LuaTypeToGoType(L, luaValue)
	fmt.Printf("Round-trip (Go->Lua->Go): original=%v, back=%v\n", originalGoValue, backToGo)

	// Test edge cases
	fmt.Println("\n=== Testing Edge Cases ===")

	// Empty array
	emptyGoArray := []any{}
	emptyLuaArray := GoTypeToLuaType(L, emptyGoArray)
	emptyBackToGo := LuaTypeToGoType(L, emptyLuaArray)
	fmt.Printf("Empty array: %v -> %v -> %v\n", emptyGoArray, emptyLuaArray, emptyBackToGo)

	// Empty map
	emptyGoMap := map[string]any{}
	emptyLuaMap := GoTypeToLuaType(L, emptyGoMap)
	emptyBackToGoMap := LuaTypeToGoType(L, emptyLuaMap)
	fmt.Printf("Empty map: %v -> %v -> %v\n", emptyGoMap, emptyLuaMap, emptyBackToGoMap)

	// Array with gaps
	gapGoArray := []any{"first", nil, "third"}
	gapLuaArray := GoTypeToLuaType(L, gapGoArray)
	gapBackToGo := LuaTypeToGoType(L, gapLuaArray)
	fmt.Printf("Array with gaps: %v -> %v -> %v\n", gapGoArray, gapLuaArray, gapBackToGo)

	fmt.Println("\n=== All tests completed successfully! ===")

	return nil
}

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
	case *lua.LFunction:
		// Functions are not directly convertible to Go types
		return nil
	case *lua.LUserData:
		// UserData is not directly convertible
		return nil
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

type MyUserData struct {
	Name   string         `json:"name"`
	Age    int            `json:"age"`
	Colors []string       `json:"colors"`
	Params map[string]any `json:"params"`
}

func MapToStruct(l *lua.LState, lvalue lua.LValue, target any) error {
	// impl

	return nil
}
