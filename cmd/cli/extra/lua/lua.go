package luaextra

import (
	"errors"
	"fmt"
	"os"

	"github.com/blue-monads/potatoverse/cmd/cli"
)

func init() {
	cli.RegisterExtraCommand("lua", RunLua)
}

// RunLua is the dispatcher for "potatoverse extra lua".
func RunLua(args []string) error {
	if len(args) == 0 {
		return RunShell()
	}

	subcommand := args[0]
	switch subcommand {
	case "shell":
		return RunShell()

	case "check":
		return RunCheck(args[1:])

	case "-e", "--eval":
		if len(args) < 2 {
			return errors.New("must provide code to evaluate: potatoverse extra lua -e <code>")
		}
		L := NewShellState()
		defer L.Close()
		return L.DoString(args[1])

	case "-h", "--help", "help":
		printHelp()
		return nil

	default:
		// If it's a file, execute it
		if fi, err := os.Stat(subcommand); err == nil && !fi.IsDir() {
			L := NewShellState()
			defer L.Close()
			return L.DoFile(subcommand)
		}

		printHelp()
		return fmt.Errorf("unknown lua command: %s", subcommand)
	}
}

func printHelp() {
	fmt.Println(`Usage: potatoverse extra lua [command] [options]

Commands:
  shell               Start interactive Lua shell (default)
  check <file.lua...> Check syntax of one or more Lua files
  -e <code>           Evaluate inline Lua code
  <file.lua>          Execute a Lua script file
  help                Show this help message

Examples:
  potatoverse extra lua
  potatoverse extra lua shell
  potatoverse extra lua check server.lua
  potatoverse extra lua -e 'print("hello world")'
  potatoverse extra lua script.lua`)
}
