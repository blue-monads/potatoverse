package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/blue-monads/turnix/backend/utils/qq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	db, err := sql.Open("sqlite3", "file:test.db?_journal_mode=WAL&_synchronous=NORMAL&_timeout=5000&vfs=vfsync")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	qq.Println("Database opened")

	_, err = db.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		log.Fatal(err)
	}

	qq.Println("Journal mode set to WAL")

	_, err = db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO test (name) VALUES (?)", "test")
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("SELECT * FROM test")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, name)
	}

	// let batch insert 10 records in txn
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	for i := range 10 {
		_, err = tx.Exec("INSERT INTO test (name) VALUES (?)", fmt.Sprintf("test%d", i))
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

}
