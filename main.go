package main

import (
	"github.com/blue-monads/potatoverse/cmd/cli"

	_ "github.com/blue-monads/potatoverse/backend/services/datahub/provider/ncruces"
	//_ "github.com/blue-monads/potatoverse/backend/services/datahub/provider/mattn"

	_ "github.com/blue-monads/potatoverse/backend/distro"
)

func main() {

	cli.Run()

}
