package cli

import (
	"fmt"
	"sync"

	"github.com/alecthomas/kong"
)

var (
	extraCommans = map[string]func(args []string) error{}
	ecLock       sync.Mutex
)

func RegisterExtraCommand(name string, runner func(args []string) error) {
	ecLock.Lock()
	defer ecLock.Unlock()

	extraCommans[name] = runner
}

func GetExtraCommand(name string) func(args []string) error {
	ecLock.Lock()
	defer ecLock.Unlock()

	runner, found := extraCommans[name]
	if !found {
		return nil
	}
	return runner
}

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

	runner := GetExtraCommand(e.Command)
	if runner == nil {
		return fmt.Errorf("extra command not found: %s", e.Command)
	}

	return runner(e.Args)

}
