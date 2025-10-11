package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

// CLI represents the overall command structure.
type CLI struct {
	// The 'exec' command demonstrates Positional Passthrough.
	Exec ExecCmd `cmd:"" help:"[MODE 1] Execute a binary, passing remaining arguments unparsed (like 'docker exec')."`

	// The 'cmd' command demonstrates Command Passthrough.
	Cmd CmdCmd `cmd:"" help:"[MODE 2] Capture ALL arguments following 'cmd' unparsed."`

	// Global flag to demonstrate that global flags are still parsed correctly.
	Verbose bool `name:"verbose" short:"v" help:"Enable verbose output."`
}

// --- MODE 1: Positional Argument Passthrough (passthrough:"partial") ---

// ExecCmd uses passthrough:"partial" on a single []string to capture everything after
// its own flags, preventing Kong from parsing any of those subsequent tokens.
type ExecCmd struct {
	// A sample flag for the 'exec' command itself.
	DryRun bool `help:"Simulate the execution without running."`

	// CombinedArgs captures ALL positional tokens after the flags.
	// passthrough:"partial" stops flag parsing when the first token for this
	// argument is encountered. It will capture the binary name and its args.
	CombinedArgs []string `arg:"" passthrough:"partial" help:"Binary and its arguments. Will not be parsed as flags."`

	// Internal fields populated in Run() for clean output
	Binary string
	Args   []string
}

// Run executes the ExecCmd.
func (e *ExecCmd) Run(ctx *kong.Context) error {
	fmt.Printf("\n--- [MODE 1] Positional Passthrough (passthrough:\"partial\") ---\n")

	// Access the parent CLI structure via the context's target.
	parentCLI := ctx.Model.Parent.Target.Interface().(*CLI)
	if parentCLI.Verbose {
		fmt.Printf("[Global Flag] Verbose: true\n")
	}

	if len(e.CombinedArgs) == 0 {
		return fmt.Errorf("must specify binary and arguments (e.g., 'ls -l')")
	}

	// Manually split the captured arguments into the Binary name and its Args
	e.Binary = e.CombinedArgs[0]
	if len(e.CombinedArgs) > 1 {
		e.Args = e.CombinedArgs[1:]
	} else {
		e.Args = []string{}
	}

	fmt.Printf("[Command Flag] DryRun: %t\n", e.DryRun)
	fmt.Printf("Executing: %s\n", e.Binary)
	fmt.Printf("With Unparsed Args: %v\n", e.Args)

	return nil
}

// --- MODE 2: Command Passthrough (passthrough:"all") ---

// CmdCmd demonstrates command passthrough. When this is set on a command,
// the command MUST contain only one argument of type []string.
type CmdCmd struct {
	// Args is the single argument that captures everything following the command.
	// passthrough:"all" causes ALL tokens after the command name to be collected
	// into this []string, regardless of whether they look like flags.
	Args []string `arg:"" passthrough:"all" help:"All arguments following this command, unparsed."`
}

// Run executes the CmdCmd.
func (c *CmdCmd) Run(ctx *kong.Context) error {
	fmt.Printf("\n--- [MODE 2] Command Passthrough (passthrough:\"all\") ---\n")

	// Access the parent CLI structure via the context's target.
	parentCLI := ctx.Model.Parent.Target.Interface().(*CLI)
	if parentCLI.Verbose {
		fmt.Printf("[Global Flag] Verbose: true\n")
	}

	fmt.Printf("Captured ALL Unparsed Args: %v\n", c.Args)
	fmt.Printf("Note: Everything after 'cmd' is captured, including '--foo' which is *not* parsed as a flag.\n")
	return nil
}

func main() {
	var cli CLI
	parser := kong.Must(&cli,
		kong.Name("mycli"),
		kong.Description("Demonstrates Kong's passthrough features."),
		kong.UsageOnError(),
	)

	// We use os.Args[1:] here, allowing us to manually simulate running the CLI.
	_, err := parser.Parse(os.Args[1:])
	if err != nil {
		if len(os.Args) == 1 {
			fmt.Println(parser.Model.Help)
			os.Exit(0)
		}

		parser.FatalIfErrorf(err)

	}

}

/*
To test this:

1. Save the code as main.go and run: go run main.go

2. Test Positional Passthrough ('exec' command) - using passthrough:"partial"
   - Flags for 'mycli' and 'exec' are parsed, but everything after 'exec' flags is unparsed.
     go run main.go -v exec --dry-run ls -l -a --user=root

   Expected Output (now correctly captures -l):
   --- [MODE 1] Positional Passthrough (passthrough:"partial") ---
   [Global Flag] Verbose: true
   [Command Flag] DryRun: true
   Executing: ls
   With Unparsed Args: [-l -a --user=root]

3. Test Command Passthrough ('cmd' command) - using passthrough:"all"
   - Only global flags (-v) are parsed. Everything after 'cmd' is captured verbatim:
     go run main.go -v cmd my-param --port 8080 -f file.txt

   Expected Output:
   --- [MODE 2] Command Passthrough (passthrough:"all") ---
   [Global Flag] Verbose: true
   Captured ALL Unparsed Args: [my-param --port 8080 -f file.txt]
   Note: Everything after 'cmd' is captured, including '--foo' which is *not* parsed as a flag.

*/
