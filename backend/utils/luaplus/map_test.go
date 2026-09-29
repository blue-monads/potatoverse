package luaplus

import (
	"encoding/json"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func TestMapToTableNumbers(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	input := map[string]any{
		"account_name": "sales",
		"id":           int64(2),
		"count":        int32(10),
		"standard_int": int(42),
		"ratio":        float64(3.14),
		"float_num":    float32(1.5),
		"unsigned":     uint64(100),
		"flag":         true,
	}

	tbl := MapToTable(L, input)

	// Check id is LTNumber in Lua table
	idVal := tbl.RawGetString("id")
	if idVal.Type() != lua.LTNumber {
		t.Fatalf("expected id to be LTNumber, got %s (%s)", idVal.Type(), idVal.String())
	}
	if float64(idVal.(lua.LNumber)) != 2 {
		t.Fatalf("expected id to be 2, got %v", idVal)
	}

	// Check round-trip through LuaToAny and json.Marshal
	converted := LuaToAny(L, tbl)
	jsonBytes, err := json.Marshal(converted)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// In unmarshaled JSON, numbers become float64
	if v, ok := parsed["id"].(float64); !ok || v != 2 {
		t.Errorf("expected id in JSON to be number 2, got %T (%v)", parsed["id"], parsed["id"])
	}

	if v, ok := parsed["account_name"].(string); !ok || v != "sales" {
		t.Errorf("expected account_name to be 'sales', got %T (%v)", parsed["account_name"], parsed["account_name"])
	}

	if v, ok := parsed["count"].(float64); !ok || v != 10 {
		t.Errorf("expected count to be number 10, got %T (%v)", parsed["count"], parsed["count"])
	}
}

func TestMapToTableNestedAndSlices(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	input := map[string]any{
		"items": []any{
			map[string]any{
				"id":    int64(1),
				"title": "item 1",
			},
			map[string]any{
				"id":    int64(2),
				"title": "item 2",
			},
		},
	}

	tbl := MapToTable(L, input)
	converted := LuaToAny(L, tbl)

	jsonBytes, err := json.Marshal(converted)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	items, ok := parsed["items"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("expected items to be array of len 2, got %v", parsed["items"])
	}

	first := items[0].(map[string]any)
	if v, ok := first["id"].(float64); !ok || v != 1 {
		t.Errorf("expected first item id to be number 1, got %T (%v)", first["id"], first["id"])
	}
}

func TestArrayOfObjectsDataShape(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	luaScript := `
		data = {
			{
				id = 1,
				name = "Test 1",
				description = "We are the test 1",
				created_at = "2021-01-01",
				updated_at = "2021-01-01"
			},
			{
				id = 2,
				name = "Test 2",
				description = "We are the test 2",
				created_at = "2021-01-01",
				updated_at = "2021-01-01"
			}
		}
	`
	if err := L.DoString(luaScript); err != nil {
		t.Fatalf("Lua DoString failed: %v", err)
	}

	dataVal := L.GetGlobal("data")

	// 1. Test LuaToAny (used by req.json)
	converted := LuaToAny(L, dataVal)
	jsonBytes, err := json.Marshal(converted)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	t.Logf("LuaToAny -> JSON: %s", string(jsonBytes))

	// Verify JSON is an array, NOT an object
	if len(jsonBytes) == 0 || jsonBytes[0] != '[' {
		t.Fatalf("Expected JSON to start with '[', but got: %s", string(jsonBytes))
	}

	var parsed []map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("json.Unmarshal into []map[string]any failed: %v", err)
	}

	if len(parsed) != 2 {
		t.Fatalf("expected 2 items, got %d", len(parsed))
	}

	if parsed[0]["id"] != float64(1) || parsed[0]["name"] != "Test 1" {
		t.Errorf("unexpected parsed[0]: %+v", parsed[0])
	}
	if parsed[1]["id"] != float64(2) || parsed[1]["name"] != "Test 2" {
		t.Errorf("unexpected parsed[1]: %+v", parsed[1])
	}

	// 2. Test TableToArray (used by req.json_array)
	arrConverted := TableToArray(L, dataVal.(*lua.LTable))
	arrJsonBytes, err := json.Marshal(arrConverted)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	t.Logf("TableToArray -> JSON: %s", string(arrJsonBytes))
	if len(arrJsonBytes) == 0 || arrJsonBytes[0] != '[' {
		t.Fatalf("Expected JSON to start with '[', but got: %s", string(arrJsonBytes))
	}
}

func TestGoSliceOfMapsRoundTrip(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Typical shape returned from database query or Go handlers
	data := []map[string]any{
		{
			"id":          int64(1),
			"name":        "Test 1",
			"description": "We are the test 1",
			"created_at":  "2021-01-01",
			"updated_at":  "2021-01-01",
		},
		{
			"id":          int64(2),
			"name":        "Test 2",
			"description": "We are the test 2",
			"created_at":  "2021-01-01",
			"updated_at":  "2021-01-01",
		},
	}

	// 1. Go -> Lua via GoTypeToLuaType
	luaVal := GoTypeToLuaType(L, data)
	if luaVal.Type() != lua.LTTable {
		t.Fatalf("expected LTTable, got %s", luaVal.Type())
	}

	// 2. Lua -> Go via LuaToAny (as in req.json)
	goVal := LuaToAny(L, luaVal)
	jsonBytes, err := json.Marshal(goVal)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	t.Logf("GoTypeToLuaType -> LuaToAny -> JSON: %s", string(jsonBytes))

	if len(jsonBytes) == 0 || jsonBytes[0] != '[' {
		t.Fatalf("Expected array JSON starting with '[', got: %s", string(jsonBytes))
	}

	var parsed []map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if len(parsed) != 2 {
		t.Fatalf("expected 2 items, got %d", len(parsed))
	}
	if parsed[0]["id"] != float64(1) {
		t.Errorf("expected id 1, got %v", parsed[0]["id"])
	}
	if parsed[1]["id"] != float64(2) {
		t.Errorf("expected id 2, got %v", parsed[1]["id"])
	}
}

func TestEmptyTableToNil(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Empty table directly
	emptyTbl := L.NewTable()
	result := LuaToAny(L, emptyTbl)
	if result != nil {
		t.Errorf("expected empty table to convert to nil, got %v (%T)", result, result)
	}

	// Table containing empty table
	containerTbl := L.NewTable()
	containerTbl.RawSetString("name", lua.LString("test"))
	containerTbl.RawSetString("empty_data", emptyTbl)

	containerResult := LuaToAny(L, containerTbl)
	jsonBytes, err := json.Marshal(containerResult)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	t.Logf("Container with empty table -> JSON: %s", string(jsonBytes))

	var parsed map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if parsed["name"] != "test" {
		t.Errorf("expected name 'test', got %v", parsed["name"])
	}
	if parsed["empty_data"] != nil {
		t.Errorf("expected empty_data to be nil/null, got %v", parsed["empty_data"])
	}
}

func TestMapToTable_Time(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	now := time.Date(2026, 9, 30, 1, 55, 0, 0, time.UTC)
	input := map[string]any{
		"last_updated": now,
		"table_id":     1,
	}

	tbl := MapToTable(L, input)
	converted := LuaToAny(L, tbl)

	jsonBytes, err := json.Marshal(converted)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expected := "2026-09-30T01:55:00Z"
	if parsed["last_updated"] != expected {
		t.Errorf("expected last_updated '%s', got '%v'", expected, parsed["last_updated"])
	}
}


