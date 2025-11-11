package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/blue-monads/turnix/backend/utils/qq"
)

func main() {

	db, err := sql.Open("sqlite3", "file:test.db?journal_mode=WAL&_journal_mode=WAL&synchronous=NORMAL&_synchronous=NORMAL&timeout=5000&_timeout=5000&vfs=vfsync")
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

	qq.Println("Txn committed")

	rows, err = db.Query("PRAGMA journal_mode")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var mode string
		err = rows.Scan(&mode)
		qq.Println("Journal mode", mode)
	}

	// 	"Unlock" 0
	// "Txn committed"
	// "Journal mode" "delete"
	// "DeviceCharacteristics"
	// "Unlock" 0
	// "Close"

}
