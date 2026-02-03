package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub/database/schema"
	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

var (
	tmpFolder string
)

func getDatabase() db.Session {
	tmpDir, err := os.MkdirTemp("", "datahub_test_*")
	if err != nil {
		panic(fmt.Errorf("failed to create temp dir: %w", err))
	}

	tmpFolder = tmpDir

	// Create a new SQLite database
	settings := sqlite.ConnectionURL{
		Database: tmpDir + "/test.db",
	}

	db, err := sqlite.Open(settings)
	if err != nil {
		panic(fmt.Errorf("failed to open database: %w", err))
	}

	sqlconn := db.Driver().(*sql.DB)

	schemaTxt := schema.Get()
	_, err = sqlconn.Exec(schemaTxt)
	if err != nil {
		panic(fmt.Errorf("failed to execute schema: %w", err))
	}

	_, err = sqlconn.Exec("CREATE TABLE IF NOT EXISTS test (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	if err != nil {
		panic(fmt.Errorf("failed to insert user: %w", err))
	}

	_, err = sqlconn.Exec("CREATE TABLE IF NOT EXISTS test2 (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)")
	if err != nil {
		panic(fmt.Errorf("failed to insert user: %w", err))
	}

	return db
}

func main() {

	defer os.RemoveAll(tmpFolder)

	db := getDatabase()
	defer db.Close()

	sqlconn := db.Driver().(*sql.DB)

	ls := lazysyncer.NewTest(lazysyncer.Options{
		DbSession:     db,
		IsSelfEnabled: true,
		Buddies:       []string{"buddy1"},
		BasePath:      tmpFolder,
	})

	_, err := sqlconn.Exec("INSERT INTO test2 (name) VALUES ('Alice')")
	if err != nil {
		panic(fmt.Errorf("failed to insert user: %w", err))
	}

	err = ls.Start()
	if err != nil {
		panic(fmt.Errorf("failed to start syncer: %w", err))
	}

	for {

		time.Sleep(10 * time.Second)

		_, err = sqlconn.Exec("INSERT INTO test (name) VALUES ('Alice')")
		if err != nil {
			panic(fmt.Errorf("failed to insert user: %w", err))
		}

	}

}
