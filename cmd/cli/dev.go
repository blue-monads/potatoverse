package cli

import "github.com/alecthomas/kong"

type DevCmd struct {
	Push DevPushCmd `cmd:"" help:"Push development changes."`
}

type DevRunStatelessCmd struct {
	Port int    `name:"port" short:"p" help:"Server port." default:"7777"`
	Host string `name:"host" help:"Server host." default:"*.localhost"`
}

func (c *DevRunStatelessCmd) Run(ctx *kong.Context) error {
	panic("Dev Run Stateless - Not implemented yet")
}

type DevPushCmd struct {
	Target string `arg:"" help:"Push target."`
}

func (c *DevPushCmd) Run(ctx *kong.Context) error {
	panic("Dev Push - Not implemented yet")

}
