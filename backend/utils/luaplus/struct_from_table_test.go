package luaplus

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

type TargetStruct struct {
	Name     string         `lua:"name"`
	Age      int            `json:"age"`
	IsAdmin  bool           `json:"is_admin"`
	Score    float64
	Tags     []string       `lua:"tags"`
	Metadata map[string]any `lua:"metadata"`
	Nested   TargetNested   `lua:"nested"`
	SkipMe   string         `lua:"-"`
	Pointer  *int           `lua:"pointer"`
}

type TargetNested struct {
	Value string `lua:"value"`
}

func TestMapToStruct(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	table := L.NewTable()
	table.RawSetString("name", lua.LString("Charlie"))
	table.RawSetString("age", lua.LNumber(42))
	table.RawSetString("is_admin", lua.LBool(true))
	table.RawSetString("score", lua.LNumber(99.5)) // tests snake_case fallback

	tagsTable := L.NewTable()
	tagsTable.Append(lua.LString("admin"))
	tagsTable.Append(lua.LString("user"))
	table.RawSetString("tags", tagsTable)

	metadataTable := L.NewTable()
	metadataTable.RawSetString("key1", lua.LString("value1"))
	table.RawSetString("metadata", metadataTable)

	nestedTable := L.NewTable()
	nestedTable.RawSetString("value", lua.LString("nested_value"))
	table.RawSetString("nested", nestedTable)

	table.RawSetString("pointer", lua.LNumber(123))

	var target TargetStruct
	err := MapToStruct(L, table, &target)
	if err != nil {
		t.Fatalf("MapToStruct failed: %v", err)
	}

	if target.Name != "Charlie" {
		t.Errorf("expected Name 'Charlie', got '%s'", target.Name)
	}
	if target.Age != 42 {
		t.Errorf("expected Age 42, got %d", target.Age)
	}
	if !target.IsAdmin {
		t.Errorf("expected IsAdmin true, got %v", target.IsAdmin)
	}
	if target.Score != 99.5 {
		t.Errorf("expected Score 99.5, got %f", target.Score)
	}

	if len(target.Tags) != 2 || target.Tags[0] != "admin" || target.Tags[1] != "user" {
		t.Errorf("expected Tags ['admin', 'user'], got %v", target.Tags)
	}

	if val, ok := target.Metadata["key1"].(string); !ok || val != "value1" {
		t.Errorf("expected Metadata['key1'] == 'value1', got %v", target.Metadata)
	}

	if target.Nested.Value != "nested_value" {
		t.Errorf("expected Nested.Value 'nested_value', got '%s'", target.Nested.Value)
	}

	if target.Pointer == nil || *target.Pointer != 123 {
		if target.Pointer == nil {
			t.Errorf("expected Pointer to not be nil")
		} else {
			t.Errorf("expected Pointer value 123, got %d", *target.Pointer)
		}
	}
}

func TestMapToStruct_Errors(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	var target TargetStruct

	// Test nil lvalue
	if err := MapToStruct(L, lua.LNil, &target); err == nil {
		t.Errorf("expected error for nil lvalue, got none")
	}

	// Test non-table lvalue
	if err := MapToStruct(L, lua.LString("test"), &target); err == nil {
		t.Errorf("expected error for non-table lvalue, got none")
	}

	// Test non-pointer target
	table := L.NewTable()
	if err := MapToStruct(L, table, target); err == nil {
		t.Errorf("expected error for non-pointer target, got none")
	}
}

func TestStructToMap(t *testing.T) {
	input := TargetStruct{
		Name:   "Dave",
		Age:    50,
		SkipMe: "should_be_skipped",
	}

	m := StructToMap(input)

	if m["name"] != "Dave" {
		t.Errorf("expected name 'Dave', got %v", m["name"])
	}
	if m["age"] != 50 {
		t.Errorf("expected age 50, got %v", m["age"])
	}
	if _, ok := m["SkipMe"]; ok {
		t.Errorf("expected SkipMe to be skipped")
	}
	if _, ok := m["-"]; ok {
		t.Errorf("expected - to be skipped")
	}
}
