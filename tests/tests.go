package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/blue-monads/turnix/backend/services/datahub/database"
)

func main() {

	// create sqlite db

	fmt.Println("@test_start")
	defer fmt.Println("@test_end")

	sdb, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer sdb.Close()

	database, err := database.FromSqlHandle(sdb)
	if err != nil {
		log.Fatalf("Failed to create database instance: %v", err)
	}

	// Use the database instance for testing
	_ = database

}
