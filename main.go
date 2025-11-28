package main

import (
	"github.com/blue-monads/turnix/cmd/cli"

	// _ "github.com/blue-monads/turnix/backend/services/datahub/provider/ncruces"
	_ "github.com/blue-monads/turnix/backend/services/datahub/provider/mattn"
)

func main() {

	cli.Run()

}
