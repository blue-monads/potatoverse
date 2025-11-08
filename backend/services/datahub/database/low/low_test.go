package low

import (
	"database/sql"
	"os"
	"testing"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) (db.Session, func()) {
	settings := sqlite.ConnectionURL{
		Database: "dbtemp.sqlite",
	}

	sess, err := sqlite.Open(settings)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create a test table
	driver := sess.Driver().(*sql.DB)
	_, err = driver.Exec(`
		CREATE TABLE IF NOT EXISTS test_table (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT,
			age INTEGER,
			active INTEGER DEFAULT 1
		)
	`)
	if err != nil {
		sess.Close()
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Ensure table is empty at start
	_, err = driver.Exec(`DELETE FROM test_table`)
	if err != nil {
		sess.Close()
		t.Fatalf("Failed to clear test table: %v", err)
	}

	cleanup := func() {
		// Clear table before closing
		driver := sess.Driver().(*sql.DB)
		driver.Exec(`DELETE FROM test_table`)
		sess.Close()

		os.Remove("dbtemp.sqlite")
	}

	return sess, cleanup
}

func TestRunDDL(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Drop table if it exists from previous test runs
	_ = ldb.RunDDL(`DROP TABLE IF EXISTS ddl_test`)

	// Test creating a new table
	err := ldb.RunDDL(`
		CREATE TABLE ddl_test (
			id INTEGER PRIMARY KEY,
			value TEXT
		)
	`)
	if err != nil {
		t.Fatalf("RunDDL failed: %v", err)
	}

	// Verify table exists by inserting data
	err = ldb.RunDDL(`INSERT INTO ddl_test (id, value) VALUES (1, 'test')`)
	if err != nil {
		t.Fatalf("RunDDL insert failed: %v", err)
	}

	// Query to verify
	result, err := ldb.RunQueryOne("SELECT * FROM ddl_test WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("RunQueryOne failed: %v", err)
	}
	if result["value"] != "test" {
		t.Errorf("Expected 'test', got %v", result["value"])
	}
}

func TestRunQuery(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Insert test data
	_, err := ldb.Insert("test_table", map[string]any{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   30,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = ldb.Insert("test_table", map[string]any{
		"name":  "Bob",
		"email": "bob@example.com",
		"age":   25,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Test query
	results, err := ldb.RunQuery("SELECT * FROM test_table ORDER BY id")
	if err != nil {
		t.Fatalf("RunQuery failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if results[0]["name"] != "Alice" {
		t.Errorf("Expected first result name to be 'Alice', got %v", results[0]["name"])
	}

	if results[1]["name"] != "Bob" {
		t.Errorf("Expected second result name to be 'Bob', got %v", results[1]["name"])
	}
}

func TestRunQueryOne(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Insert test data
	_, err := ldb.Insert("test_table", map[string]any{
		"name":  "Charlie",
		"email": "charlie@example.com",
		"age":   35,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Test query one
	result, err := ldb.RunQueryOne("SELECT * FROM test_table WHERE name = ?", "Charlie")
	if err != nil {
		t.Fatalf("RunQueryOne failed: %v", err)
	}

	if result["name"] != "Charlie" {
		t.Errorf("Expected name to be 'Charlie', got %v", result["name"])
	}

	if result["email"] != "charlie@example.com" {
		t.Errorf("Expected email to be 'charlie@example.com', got %v", result["email"])
	}

	// Test query one with no results
	_, err = ldb.RunQueryOne("SELECT * FROM test_table WHERE name = ?", "Nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent record, got nil")
	}
}

func TestInsert(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Test insert
	id, err := ldb.Insert("test_table", map[string]any{
		"name":  "David",
		"email": "david@example.com",
		"age":   28,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if id <= 0 {
		t.Errorf("Expected positive ID, got %d", id)
	}

	// Verify insert
	result, err := ldb.FindById("test_table", id)
	if err != nil {
		t.Fatalf("FindById failed: %v", err)
	}

	if result["name"] != "David" {
		t.Errorf("Expected name to be 'David', got %v", result["name"])
	}
}

func TestUpdateById(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Insert test data
	id, err := ldb.Insert("test_table", map[string]any{
		"name":  "Eve",
		"email": "eve@example.com",
		"age":   32,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Update
	err = ldb.UpdateById("test_table", id, map[string]any{
		"name": "Eve Updated",
		"age":  33,
	})
	if err != nil {
		t.Fatalf("UpdateById failed: %v", err)
	}

	// Verify update
	result, err := ldb.FindById("test_table", id)
	if err != nil {
		t.Fatalf("FindById failed: %v", err)
	}

	if result["name"] != "Eve Updated" {
		t.Errorf("Expected name to be 'Eve Updated', got %v", result["name"])
	}

	if result["age"] != int64(33) {
		t.Errorf("Expected age to be 33, got %v", result["age"])
	}

	// Email should remain unchanged
	if result["email"] != "eve@example.com" {
		t.Errorf("Expected email to remain 'eve@example.com', got %v", result["email"])
	}
}

func TestDeleteById(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Insert test data
	id, err := ldb.Insert("test_table", map[string]any{
		"name":  "Frank",
		"email": "frank@example.com",
		"age":   40,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Delete
	err = ldb.DeleteById("test_table", id)
	if err != nil {
		t.Fatalf("DeleteById failed: %v", err)
	}

	// Verify deletion
	_, err = ldb.FindById("test_table", id)
	if err == nil {
		t.Error("Expected error when finding deleted record, got nil")
	}
}

func TestFindById(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Insert test data
	id, err := ldb.Insert("test_table", map[string]any{
		"name":  "Grace",
		"email": "grace@example.com",
		"age":   27,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Find by ID
	result, err := ldb.FindById("test_table", id)
	if err != nil {
		t.Fatalf("FindById failed: %v", err)
	}

	if result["name"] != "Grace" {
		t.Errorf("Expected name to be 'Grace', got %v", result["name"])
	}

	if result["email"] != "grace@example.com" {
		t.Errorf("Expected email to be 'grace@example.com', got %v", result["email"])
	}

	// Test with nonexistent ID
	_, err = ldb.FindById("test_table", 99999)
	if err == nil {
		t.Error("Expected error for nonexistent ID, got nil")
	}
}

func TestUpdateByCond(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Insert test data
	_, err := ldb.Insert("test_table", map[string]any{
		"name":   "Henry",
		"email":  "henry@example.com",
		"age":    29,
		"active": 1,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = ldb.Insert("test_table", map[string]any{
		"name":   "Iris",
		"email":  "iris@example.com",
		"age":    31,
		"active": 1,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Update by condition
	err = ldb.UpdateByCond("test_table", map[any]any{
		"active": 1,
	}, map[string]any{
		"active": 0,
	})
	if err != nil {
		t.Fatalf("UpdateByCond failed: %v", err)
	}

	// Verify both records were updated
	results, err := ldb.FindAllByQuery(&datahub.FindQuery{
		Table:  "test_table",
		Cond:   map[any]any{"active": 0},
		Offset: 0,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestDeleteByCond(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Insert test data
	_, err := ldb.Insert("test_table", map[string]any{
		"name":  "Jack",
		"email": "jack@example.com",
		"age":   26,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = ldb.Insert("test_table", map[string]any{
		"name":  "Kate",
		"email": "kate@example.com",
		"age":   24,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Delete by condition
	err = ldb.DeleteByCond("test_table", map[any]any{
		"name": "Jack",
	})
	if err != nil {
		t.Fatalf("DeleteByCond failed: %v", err)
	}

	// Verify deletion
	results, err := ldb.FindAllByQuery(&datahub.FindQuery{
		Table:  "test_table",
		Offset: 0,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result after deletion, got %d", len(results))
	}

	if results[0]["name"] != "Kate" {
		t.Errorf("Expected remaining record to be 'Kate', got %v", results[0]["name"])
	}
}

func TestFindAll(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Insert multiple test records
	names := []string{"UserA", "UserB", "UserC", "UserD", "UserE"}
	emails := []string{"usera@example.com", "userb@example.com", "userc@example.com", "userd@example.com", "usere@example.com"}
	for i := 0; i < 5; i++ {
		_, err := ldb.Insert("test_table", map[string]any{
			"name":  names[i],
			"email": emails[i],
			"age":   20 + i,
		})
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	}

	// Test FindAll without conditions
	results, err := ldb.FindAllByCond("test_table", map[any]any{})
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}

	// Test FindAll with offset and limit
	results, err = ldb.FindAllByQuery(&datahub.FindQuery{
		Table:  "test_table",
		Offset: 1,
		Limit:  2,
	})
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results with offset 1 and limit 2, got %d", len(results))
	}

	// Test FindAll with condition
	results, err = ldb.FindAllByQuery(&datahub.FindQuery{
		Table: "test_table",
		Cond: map[any]any{
			"age": 22,
		},
		Offset: 0,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result with age=22, got %d", len(results))
	}
}

func TestFindOne(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Insert test data
	_, err := ldb.Insert("test_table", map[string]any{
		"name":  "Liam",
		"email": "liam@example.com",
		"age":   28,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Test FindOne
	result, err := ldb.FindOneByCond("test_table", map[any]any{
		"name": "Liam",
	})
	if err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}

	if result["name"] != "Liam" {
		t.Errorf("Expected name to be 'Liam', got %v", result["name"])
	}

	if result["email"] != "liam@example.com" {
		t.Errorf("Expected email to be 'liam@example.com', got %v", result["email"])
	}

	// Test FindOne with nonexistent record
	_, err = ldb.FindOneByCond("test_table", map[any]any{
		"name": "Nonexistent",
	})
	if err == nil {
		t.Error("Expected error for nonexistent record, got nil")
	}
}

func TestFindAllWithEmptyCond(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Insert test data
	_, err := ldb.Insert("test_table", map[string]any{
		"name": "Test",
		"age":  25,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Test FindAll with empty condition map
	results, err := ldb.FindAllByCond("test_table", map[any]any{})
	if err != nil {
		t.Fatalf("FindAll with empty condition failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestFindAllOffsetLimit(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Insert 10 records
	for i := 0; i < 10; i++ {
		_, err := ldb.Insert("test_table", map[string]any{
			"name": "Record" + string(rune('0'+i%10)),
			"age":  i,
		})
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
	}

	// Test offset 0, limit 5
	results, err := ldb.FindAllByQuery(&datahub.FindQuery{
		Table:  "test_table",
		Offset: 0,
		Limit:  5,
	})
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("Expected 5 results with limit 5, got %d", len(results))
	}

	// Test offset 5, limit 5
	results, err = ldb.FindAllByQuery(&datahub.FindQuery{
		Table:  "test_table",
		Offset: 5,
		Limit:  5,
	})
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("Expected 5 results with offset 5 and limit 5, got %d", len(results))
	}

	// Test offset 0, limit 0 (should return all - limit 0 means no limit in our implementation)
	results, err = ldb.FindAllByQuery(&datahub.FindQuery{
		Table:  "test_table",
		Offset: 0,
		Limit:  0,
	})
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	// With limit 0, our implementation returns all records (limit is only applied if > 0)
	if len(results) != 10 {
		t.Errorf("Expected 10 results with limit 0, got %d", len(results))
	}
}

func TestBuildCond_EmptyCondition(t *testing.T) {
	cond := buildCond(map[any]any{})

	// Should return EmptyCond (check if it's an empty db.Cond)
	dbCond, ok := cond.(db.Cond)
	if !ok {
		t.Errorf("Expected db.Cond for empty condition, got %T", cond)
		return
	}
	if len(dbCond) != 0 {
		t.Errorf("Expected empty db.Cond, got len %d", len(dbCond))
	}
}

func TestBuildCond_SimpleCondition(t *testing.T) {
	cond := buildCond(map[any]any{
		"name": "John",
		"age":  30,
	})

	// Should return db.Cond
	dbCond, ok := cond.(db.Cond)
	if !ok {
		t.Fatalf("Expected db.Cond, got %T", cond)
	}

	if dbCond["name"] != "John" {
		t.Errorf("Expected name='John', got %v", dbCond["name"])
	}
	if dbCond["age"] != 30 {
		t.Errorf("Expected age=30, got %v", dbCond["age"])
	}
}

func TestBuildCond_ANDCondition(t *testing.T) {
	cond := buildCond(map[any]any{
		"AND": []any{
			map[any]any{"name": "John"},
			map[any]any{"age": 30},
		},
	})

	// Should return db.LogicalExpr (db.And)
	_, ok := cond.(db.LogicalExpr)
	if !ok {
		t.Fatalf("Expected db.LogicalExpr for AND condition, got %T", cond)
	}
}

func TestBuildCond_ORCondition(t *testing.T) {
	cond := buildCond(map[any]any{
		"OR": []any{
			map[any]any{"name": "John"},
			map[any]any{"name": "Jane"},
		},
	})

	// Should return db.LogicalExpr
	_, ok := cond.(db.LogicalExpr)
	if !ok {
		t.Fatalf("Expected db.LogicalExpr for OR condition, got %T", cond)
	}
}

func TestBuildCond_NestedANDCondition(t *testing.T) {
	cond := buildCond(map[any]any{
		"AND": []any{
			map[any]any{
				"AND": []any{
					map[any]any{"age >": 21},
					map[any]any{"age <": 28},
				},
			},
			map[any]any{"name": "John"},
		},
	})

	// Should return db.LogicalExpr
	_, ok := cond.(db.LogicalExpr)
	if !ok {
		t.Fatalf("Expected db.LogicalExpr for nested AND condition, got %T", cond)
	}
}

func TestBuildCond_NestedORCondition(t *testing.T) {
	cond := buildCond(map[any]any{
		"OR": []any{
			map[any]any{
				"OR": []any{
					map[any]any{"name": "John"},
					map[any]any{"name": "Jane"},
				},
			},
			map[any]any{"age": 30},
		},
	})

	// Should return db.LogicalExpr
	_, ok := cond.(db.LogicalExpr)
	if !ok {
		t.Fatalf("Expected db.LogicalExpr for nested OR condition, got %T", cond)
	}
}

func TestBuildCond_MixedNestedCondition(t *testing.T) {
	cond := buildCond(map[any]any{
		"AND": []any{
			map[any]any{
				"AND": []any{
					map[any]any{"age >": 21},
					map[any]any{"age <": 28},
				},
			},
			map[any]any{
				"OR": []any{
					map[any]any{"name": "Joanna"},
					map[any]any{"name": "John"},
					map[any]any{"name": "Jhon"},
				},
			},
		},
	})

	// Should return db.LogicalExpr
	_, ok := cond.(db.LogicalExpr)
	if !ok {
		t.Fatalf("Expected db.LogicalExpr for mixed nested condition, got %T", cond)
	}
}

func TestBuildCond_WithDatabaseQuery_SimpleAND(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Insert test data
	_, err := ldb.Insert("test_table", map[string]any{
		"name":   "Alice",
		"email":  "alice@example.com",
		"age":    25,
		"active": 1,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = ldb.Insert("test_table", map[string]any{
		"name":   "Bob",
		"email":  "bob@example.com",
		"age":    30,
		"active": 1,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = ldb.Insert("test_table", map[string]any{
		"name":   "Charlie",
		"email":  "charlie@example.com",
		"age":    25,
		"active": 0,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Test AND condition using buildCond
	cond := buildCond(map[any]any{
		"AND": []any{
			map[any]any{"age": 25},
			map[any]any{"active": 1},
		},
	})

	// Use the condition in a query
	collection := sess.Collection("test_table")
	var results []map[string]any
	err = collection.Find(cond).All(&results)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Should find only Alice (age=25 AND active=1)
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if len(results) > 0 && results[0]["name"] != "Alice" {
		t.Errorf("Expected name='Alice', got %v", results[0]["name"])
	}
}

func TestBuildCond_WithDatabaseQuery_SimpleOR(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Insert test data
	_, err := ldb.Insert("test_table", map[string]any{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   25,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = ldb.Insert("test_table", map[string]any{
		"name":  "Bob",
		"email": "bob@example.com",
		"age":   30,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = ldb.Insert("test_table", map[string]any{
		"name":  "Charlie",
		"email": "charlie@example.com",
		"age":   35,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Test OR condition using buildCond
	cond := buildCond(map[any]any{
		"OR": []any{
			map[any]any{"name": "Alice"},
			map[any]any{"name": "Bob"},
		},
	})

	// Use the condition in a query
	collection := sess.Collection("test_table")
	var results []map[string]any
	err = collection.Find(cond).All(&results)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Should find Alice OR Bob (2 results)
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	names := make(map[string]bool)
	for _, r := range results {
		names[r["name"].(string)] = true
	}
	if !names["Alice"] || !names["Bob"] {
		t.Errorf("Expected Alice and Bob, got names: %v", names)
	}
}

func TestBuildCond_WithDatabaseQuery_NestedAND(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Insert test data
	_, err := ldb.Insert("test_table", map[string]any{
		"name":   "John",
		"email":  "john@example.com",
		"age":    25,
		"active": 1,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = ldb.Insert("test_table", map[string]any{
		"name":   "Jane",
		"email":  "jane@example.com",
		"age":    25,
		"active": 0,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Test nested AND condition
	cond := buildCond(map[any]any{
		"AND": []any{
			map[any]any{
				"AND": []any{
					map[any]any{"age": 25},
					map[any]any{"active": 1},
				},
			},
			map[any]any{"name": "John"},
		},
	})

	// Use the condition in a query
	collection := sess.Collection("test_table")
	var results []map[string]any
	err = collection.Find(cond).All(&results)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Should find only John (age=25 AND active=1 AND name=John)
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
	if len(results) > 0 && results[0]["name"] != "John" {
		t.Errorf("Expected name='John', got %v", results[0]["name"])
	}
}

func TestBuildCond_WithDatabaseQuery_ComplexNested(t *testing.T) {
	sess, cleanup := setupTestDB(t)
	defer cleanup()

	ldb := NewLowDB(sess, "package", "test123")

	// Insert test data matching the comment example
	_, err := ldb.Insert("test_table", map[string]any{
		"name":  "Joanna",
		"email": "joanna@example.com",
		"age":   25,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = ldb.Insert("test_table", map[string]any{
		"name":  "John",
		"email": "john@example.com",
		"age":   30,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = ldb.Insert("test_table", map[string]any{
		"name":  "Jhon",
		"email": "jhon@example.com",
		"age":   22,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	_, err = ldb.Insert("test_table", map[string]any{
		"name":  "Other",
		"email": "other@example.com",
		"age":   20,
	})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Test complex nested condition: (age > 21 AND age < 28) AND (name = "Joanna" OR name = "John" OR name = "Jhon")
	cond := buildCond(map[any]any{
		"AND": []any{
			map[any]any{
				"AND": []any{
					map[any]any{"age >": 21},
					map[any]any{"age <": 28},
				},
			},
			map[any]any{
				"OR": []any{
					map[any]any{"name": "Joanna"},
					map[any]any{"name": "John"},
					map[any]any{"name": "Jhon"},
				},
			},
		},
	})

	// Use the condition in a query
	collection := sess.Collection("test_table")
	var results []map[string]any
	err = collection.Find(cond).All(&results)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Should find Joanna (age=25, matches both conditions) and Jhon (age=22, matches both)
	// John has age=30 which doesn't match age < 28
	if len(results) < 1 {
		t.Errorf("Expected at least 1 result, got %d", len(results))
	}

	names := make(map[string]bool)
	for _, r := range results {
		names[r["name"].(string)] = true
	}

	// Joanna (age=25) should be found
	if !names["Joanna"] {
		t.Errorf("Expected to find Joanna, got names: %v", names)
	}
	// Jhon (age=22) should be found
	if !names["Jhon"] {
		t.Errorf("Expected to find Jhon, got names: %v", names)
	}
}

func TestTransformNestedCond_DepthLimit(t *testing.T) {
	// Create a deeply nested condition that exceeds depth limit (11 levels deep)
	// We'll build it programmatically to avoid syntax errors
	var buildDeepNested func(int) any
	buildDeepNested = func(depth int) any {
		if depth >= 11 {
			return map[any]any{"name": "John"}
		}
		return map[any]any{
			"AND": []any{
				buildDeepNested(depth + 1),
			},
		}
	}

	nested := []any{
		buildDeepNested(0),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for depth limit exceeded, but no panic occurred")
		} else if r != "depth limit reached" {
			t.Errorf("Expected 'depth limit reached' panic, got %v", r)
		}
	}()

	transformNestedCond(nested, 0, true)
}

func TestBuildCond_WithComparisonOperators(t *testing.T) {
	cond := buildCond(map[any]any{
		"age >": 21,
		"age <": 28,
	})

	// Should return db.Cond with comparison operators
	dbCond, ok := cond.(db.Cond)
	if !ok {
		t.Fatalf("Expected db.Cond, got %T", cond)
	}

	if dbCond["age >"] != 21 {
		t.Errorf("Expected age > 21, got %v", dbCond["age >"])
	}
	if dbCond["age <"] != 28 {
		t.Errorf("Expected age < 28, got %v", dbCond["age <"])
	}
}
