package main

import (
	"log"
	"time"

	buddyfs "github.com/blue-monads/turnix/backend/labs/buddyFs"
	"github.com/gin-gonic/gin"
)

func main() {

	buddyFs, err := buddyfs.NewBuddyFs("./tmp/buddyfs")
	if err != nil {
		log.Printf("Failed to create BuddyFs: %v", err)
	}

	engine := gin.Default()

	buddyFs.Mount(engine.Group("/buddyfs"))

	go buddyClientRun()

	engine.Run(":8666")

}

func buddyClientRun() {

	time.Sleep(2 * time.Second)

	client := buddyfs.NewBuddyFsClient("http://localhost:8666/buddyfs")

	err := client.Ping()
	if err != nil {
		log.Printf("Failed to ping: %v", err)
		return
	}

	file, err := client.Create("test.txt")
	if err != nil {
		log.Printf("Failed to create file: %v", err)
		return
	}

	defer file.Close()

	_, err = file.WriteString("Hello, World!")
	if err != nil {
		log.Printf("Failed to write to file: %v", err)
		return
	}

	file, err = client.Open("test.txt")
	if err != nil {
		log.Printf("Failed to open file: %v", err)
		return
	}
	defer file.Close()

}
