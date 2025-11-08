package cli

import (
	"fmt"

	"github.com/alecthomas/kong"
)

type ExtraCmd struct {
	CombinedArgs []string `arg:"" passthrough:"partial" help:"Extra command and its arguments."`
	Command      string
	Args         []string
}

func (e *ExtraCmd) Run(ctx *kong.Context) error {
	if len(e.CombinedArgs) == 0 {
		return fmt.Errorf("must specify command and arguments for extra passthrough")
	}

	e.Command = e.CombinedArgs[0]
	if len(e.CombinedArgs) > 1 {
		e.Args = e.CombinedArgs[1:]
	} else {
		e.Args = []string{}
	}

	fmt.Printf("Executing Extra Command: %s\n", e.Command)
	fmt.Printf("With Unparsed Args: %v\n", e.Args)

	return nil
}
