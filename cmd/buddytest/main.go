package main

import (
	"log"

	buddyfs "github.com/blue-monads/turnix/backend/labs/buddyFs"
	"github.com/gin-gonic/gin"
)

func main() {

}

func Mmain2() {

	buddyFs, err := buddyfs.NewBuddyFs("./tmp/buddyfs")
	if err != nil {
		log.Printf("Failed to create BuddyFs: %v", err)
	}

	engine := gin.Default()

	buddyFs.Mount(engine.Group("/buddyfs"))

	engine.Run(":8666")

}
