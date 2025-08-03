package main

import (
	"github.com/blue-monads/turnix/cmd/cli"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	cli.Run()

}
