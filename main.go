package main

import (
	"github.com/blue-monads/turnix/cmd/cli"
	_ "github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func main() {

	cli.Run()

}
