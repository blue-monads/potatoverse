package stateless

import (
	"fmt"
	"log"

	"github.com/blue-monads/turnix/backend/utils/luaplus"
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

type MyUserData struct {
	Name   string         `json:"name"`
	Age    int            `json:"age"`
	Colors []string       `json:"colors"`
	Params map[string]any `json:"params"`
}

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
	goNil := luaplus.LuaTypeToGoType(L, testNil)
	fmt.Printf("Lua nil -> Go: %v (type: %T)\n", goNil, goNil)

	// Test bool
	testBool := L.GetGlobal("test_bool")
	goBool := luaplus.LuaTypeToGoType(L, testBool)
	fmt.Printf("Lua bool -> Go: %v (type: %T)\n", goBool, goBool)

	// Test string
	testString := L.GetGlobal("test_string")
	goString := luaplus.LuaTypeToGoType(L, testString)
	fmt.Printf("Lua string -> Go: %v (type: %T)\n", goString, goString)

	// Test number (int)
	testNumberInt := L.GetGlobal("test_number_int")
	goNumberInt := luaplus.LuaTypeToGoType(L, testNumberInt)
	fmt.Printf("Lua number (int) -> Go: %v (type: %T)\n", goNumberInt, goNumberInt)

	// Test number (float)
	testNumberFloat := L.GetGlobal("test_number_float")
	goNumberFloat := luaplus.LuaTypeToGoType(L, testNumberFloat)
	fmt.Printf("Lua number (float) -> Go: %v (type: %T)\n", goNumberFloat, goNumberFloat)

	// Test array conversion
	fmt.Println("\n=== Testing Array Conversion ===")
	testArray := L.GetGlobal("test_array")
	goArray := luaplus.LuaTypeToGoType(L, testArray)
	fmt.Printf("Lua array -> Go: %v (type: %T)\n", goArray, goArray)

	// Test table conversion
	fmt.Println("\n=== Testing Table Conversion ===")
	testTable := L.GetGlobal("test_table")
	goTable := luaplus.LuaTypeToGoType(L, testTable)
	fmt.Printf("Lua table -> Go: %v (type: %T)\n", goTable, goTable)

	// Test nested structure conversion
	fmt.Println("\n=== Testing Nested Structure Conversion ===")
	testNested := L.GetGlobal("test_nested")
	goNested := luaplus.LuaTypeToGoType(L, testNested)
	fmt.Printf("Lua nested -> Go: %v (type: %T)\n", goNested, goNested)

	// Test mixed array/table
	fmt.Println("\n=== Testing Mixed Array/Table Conversion ===")
	testMixed := L.GetGlobal("test_mixed")
	goMixed := luaplus.LuaTypeToGoType(L, testMixed)
	fmt.Printf("Lua mixed -> Go: %v (type: %T)\n", goMixed, goMixed)

	// Test Go to Lua conversions
	fmt.Println("\n=== Testing Go to Lua Conversions ===")

	// Test Go array to Lua
	goTestArray := []any{1, 2, 3, "four", true}
	luaArray := luaplus.GoTypeToLuaType(L, goTestArray)
	fmt.Printf("Go array -> Lua: %v\n", luaArray)

	// Test Go map to Lua
	goTestMap := map[string]any{
		"name":   "Bob",
		"age":    25,
		"active": false,
		"score":  87.3,
	}
	luaMap := luaplus.GoTypeToLuaType(L, goTestMap)
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
	luaNested := luaplus.GoTypeToLuaType(L, goTestNested)
	fmt.Printf("Go nested -> Lua: %v\n", luaNested)

	// Test round-trip conversions
	fmt.Println("\n=== Testing Round-trip Conversions ===")

	// Round-trip: Lua -> Go -> Lua
	originalLuaValue := L.GetGlobal("test_table")
	goValue := luaplus.LuaTypeToGoType(L, originalLuaValue)
	backToLua := luaplus.GoTypeToLuaType(L, goValue)
	fmt.Printf("Round-trip (Lua->Go->Lua): original=%v, back=%v\n", originalLuaValue, backToLua)

	// Round-trip: Go -> Lua -> Go
	originalGoValue := []any{"test", 42, true}
	luaValue := luaplus.GoTypeToLuaType(L, originalGoValue)
	backToGo := luaplus.LuaTypeToGoType(L, luaValue)
	fmt.Printf("Round-trip (Go->Lua->Go): original=%v, back=%v\n", originalGoValue, backToGo)

	// Test struct conversions
	fmt.Println("\n=== Testing Struct Conversions ===")

	// Test Lua table to Go struct
	testStruct := &MyUserData{}
	err = luaplus.MapToStruct(L, L.GetGlobal("test_nested"), testStruct)
	if err != nil {
		fmt.Printf("Error converting Lua table to struct: %v\n", err)
	} else {
		fmt.Printf("Lua table -> Go struct: %+v\n", testStruct)
	}

	// Test Go struct to Lua table
	luaTable, err := luaplus.StructToTable(L, testStruct)
	if err != nil {
		fmt.Printf("Error converting Go struct to Lua table: %v\n", err)
	} else {
		fmt.Printf("Go struct -> Lua table: %v\n", luaTable)
	}

	// Test round-trip struct conversion
	fmt.Println("\n=== Testing Struct Round-trip ===")

	// Create a test struct
	originalStruct := &MyUserData{
		Name:   "Test User",
		Age:    25,
		Colors: []string{"red", "blue", "green"},
		Params: map[string]any{
			"theme": "dark",
			"lang":  "en",
		},
	}

	// Convert to Lua table
	luaTableFromStruct, err := luaplus.StructToTable(L, originalStruct)
	if err != nil {
		fmt.Printf("Error converting struct to Lua: %v\n", err)
	} else {
		fmt.Printf("Original struct: %+v\n", originalStruct)
		fmt.Printf("Struct -> Lua table: %v\n", luaTableFromStruct)

		// Convert back to struct
		convertedStruct := &MyUserData{}
		err = luaplus.MapToStruct(L, luaTableFromStruct, convertedStruct)
		if err != nil {
			fmt.Printf("Error converting Lua table back to struct: %v\n", err)
		} else {
			fmt.Printf("Lua table -> Struct: %+v\n", convertedStruct)
		}
	}

	// Test edge cases
	fmt.Println("\n=== Testing Edge Cases ===")

	// Empty array
	emptyGoArray := []any{}
	emptyLuaArray := luaplus.GoTypeToLuaType(L, emptyGoArray)
	emptyBackToGo := luaplus.LuaTypeToGoType(L, emptyLuaArray)
	fmt.Printf("Empty array: %v -> %v -> %v\n", emptyGoArray, emptyLuaArray, emptyBackToGo)

	// Empty map
	emptyGoMap := map[string]any{}
	emptyLuaMap := luaplus.GoTypeToLuaType(L, emptyGoMap)
	emptyBackToGoMap := luaplus.LuaTypeToGoType(L, emptyLuaMap)
	fmt.Printf("Empty map: %v -> %v -> %v\n", emptyGoMap, emptyLuaMap, emptyBackToGoMap)

	// Array with gaps
	gapGoArray := []any{"first", nil, "third"}
	gapLuaArray := luaplus.GoTypeToLuaType(L, gapGoArray)
	gapBackToGo := luaplus.LuaTypeToGoType(L, gapLuaArray)
	fmt.Printf("Array with gaps: %v -> %v -> %v\n", gapGoArray, gapLuaArray, gapBackToGo)

	fmt.Println("\n=== All tests completed successfully! ===")

	return nil
}
