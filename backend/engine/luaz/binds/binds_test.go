package binds

import (
	"database/sql"
	"testing"

	"github.com/k0kubun/pp"
	_ "github.com/mattn/go-sqlite3"
	lua "github.com/yuin/gopher-lua"
)

func TestBinds(t *testing.T) {
	pp.Println("test")

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}

	bindDb := BluaDb{
		db: db,
	}

	_, err = db.Exec(`

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);

	INSERT INTO users (name) VALUES ('John Doe');

	`)

	if err != nil {
		t.Fatalf("db.Exec(): %q\n", err)
	}

	L := lua.NewState()

	// Preload the db module first
	L.PreloadModule("db", func(L *lua.LState) int {
		return bindDb.Bind(L)
	})

	// Now load and execute the Lua code
	err = L.DoString(`
	
	function test()
		local db = require("db")

		print(db.query)

		if not db then
			error("Failed to load db module")
		end
		if not db.query then
			error("db.query is not a function")
		end
		local result, err = db.query("SELECT * FROM users")
		print(result, err)
		return result
	end

	`)
	if err != nil {
		t.Fatalf("L.DoString(): %q\n", err)
	}

	tfunc := L.GetGlobal("test")
	if tfunc.Type() != lua.LTFunction {
		t.Fatalf("Expected function, got %s", tfunc.Type())
	}

	L.Push(tfunc)
	err = L.PCall(0, 1, nil)
	if err != nil {
		t.Fatalf("Error calling test(): %v", err)
	}

	pp.Println(L.Get(-1))

}
