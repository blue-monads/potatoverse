package luaplus

import (
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type TestStruct struct {
	Name    string         `json:"name"`
	Age     int            `json:"age"`
	Active  bool           `json:"active"`
	Ignored string         `json:"-"`
	Nested  TestNested     `json:"nested"`
	Tags    []string       `json:"tags"`
	Scores  map[string]int `json:"scores"`
}

type TestNested struct {
	Value string `json:"value"`
}

func TestStructToTable(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	input := TestStruct{
		Name:    "Alice",
		Age:     30,
		Active:  true,
		Ignored: "ignore_me",
		Nested: TestNested{
			Value: "nested_value",
		},
		Tags: []string{"go", "lua"},
		Scores: map[string]int{
			"math": 95,
		},
	}

	table, err := StructToTable(L, input)
	if err != nil {
		t.Fatalf("StructToTable failed: %v", err)
	}

	if table.Type() != lua.LTTable {
		t.Fatalf("expected table, got %v", table.Type())
	}

	// Verify Name
	if name := table.RawGetString("name"); name.String() != "Alice" {
		t.Errorf("expected name 'Alice', got '%s'", name)
	}

	// Verify Age
	if age := table.RawGetString("age"); age != lua.LNumber(30) {
		t.Errorf("expected age 30, got '%s'", age)
	}

	// Verify Active
	if active := table.RawGetString("active"); active != lua.LBool(true) {
		t.Errorf("expected active true, got '%s'", active)
	}

	// Verify Ignored
	if ignored := table.RawGetString("Ignored"); ignored != lua.LNil {
		t.Errorf("expected Ignored to be nil, got '%s'", ignored)
	}

	// Verify Nested
	nested := table.RawGetString("nested")
	if nested.Type() != lua.LTTable {
		t.Errorf("expected nested to be table, got '%s'", nested.Type())
	} else {
		nestedTable := nested.(*lua.LTable)
		if val := nestedTable.RawGetString("value"); val.String() != "nested_value" {
			t.Errorf("expected nested.value 'nested_value', got '%s'", val)
		}
	}

	// Verify Tags
	tags := table.RawGetString("tags")
	if tags.Type() != lua.LTTable {
		t.Errorf("expected tags to be table, got '%s'", tags.Type())
	} else {
		tagsTable := tags.(*lua.LTable)
		if val := tagsTable.RawGetInt(1); val.String() != "go" {
			t.Errorf("expected tags[1] 'go', got '%s'", val)
		}
		if val := tagsTable.RawGetInt(2); val.String() != "lua" {
			t.Errorf("expected tags[2] 'lua', got '%s'", val)
		}
	}

	// Verify Scores
	scores := table.RawGetString("scores")
	if scores.Type() != lua.LTTable {
		t.Errorf("expected scores to be table, got '%s'", scores.Type())
	} else {
		scoresTable := scores.(*lua.LTable)
		if val := scoresTable.RawGetString("math"); val != lua.LNumber(95) {
			t.Errorf("expected scores['math'] 95, got '%s'", val)
		}
	}
}

func TestStructToTable_Pointer(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	input := &TestStruct{
		Name: "Bob",
	}

	table, err := StructToTable(L, input)
	if err != nil {
		t.Fatalf("StructToTable failed: %v", err)
	}

	if name := table.RawGetString("name"); name.String() != "Bob" {
		t.Errorf("expected name 'Bob', got '%s'", name)
	}
}

func TestStructToTable_Nil(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	table, err := StructToTable(L, nil)
	if err != nil {
		t.Fatalf("StructToTable failed: %v", err)
	}
	if table.Type() != lua.LTTable {
		t.Fatalf("expected table, got %v", table.Type())
	}
}

func TestStructToTable_Time(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	now := time.Date(2026, 9, 30, 1, 55, 0, 0, time.UTC)
	type TimeStruct struct {
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt *time.Time `json:"updated_at"`
	}

	ts := TimeStruct{
		CreatedAt: now,
		UpdatedAt: &now,
	}

	tbl, err := StructToTable(L, ts)
	if err != nil {
		t.Fatalf("StructToTable failed: %v", err)
	}

	expected := "2026-09-30T01:55:00Z"
	if ca := tbl.RawGetString("created_at").String(); ca != expected {
		t.Errorf("expected created_at '%s', got '%s'", expected, ca)
	}
	if ua := tbl.RawGetString("updated_at").String(); ua != expected {
		t.Errorf("expected updated_at '%s', got '%s'", expected, ua)
	}
}
