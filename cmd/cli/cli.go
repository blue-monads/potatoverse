package cli

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

type CLI struct {
	Server     ServerCmd     `cmd:"" help:"Server management commands."`
	Package    PackageCmd    `cmd:"" help:"Package management commands."`
	Operations OperationsCmd `cmd:"" help:"Backup and restore operations."`
	Dev        DevCmd        `cmd:"" help:"Development utilities."`
	Extra      ExtraCmd      `cmd:"" help:"Extra commands."`
	Verbose    bool          `name:"verbose" short:"v" help:"Enable verbose output."`
}

// server

type ServerCmd struct {
	Init  ServerInitCmd  `cmd:"" help:"Initialize the server with default options."`
	Start ServerStartCmd `cmd:"" help:"Start the server."`
	Stop  ServerStopCmd  `cmd:"" help:"Stop the server."`
}

type ServerInitCmd struct {
	DBFile          string `name:"db" help:"Path to database file." default:"data.db"`
	Port            int    `name:"port" short:"p" help:"Server port." default:"7777"`
	Host            string `name:"host" help:"Server host." default:"*.localhost"`
	Name            string `name:"name" help:"Name of node." default:"PotatoVerse"`
	SocketFile      string `name:"socket-file" help:"Socket file of node."`
	MasterSecret    string `name:"master-secret" help:"Master secret of node."`
	MasterSecretEnv string `name:"master-secret-env" help:"Master secret environment variable of node."`
	Debug           bool   `name:"debug" help:"Debug mode of node." default:"false"`
	WorkingDir      string `name:"working-dir" help:"Working dir of node."`
}

type ServerStartCmd struct {
	Config   string `name:"config" short:"c" help:"Path to configuration file." type:"path" default:"./config.toml"`
	AutoSeed bool   `name:"auto-seed" short:"s" help:"Auto seed the server." default:"false"`
}

// package

type PackageCmd struct {
	Build PackageBuildCmd `cmd:"" help:"Build the package."`
}

type PackageBuildCmd struct {
	PotatoTomlFile string `name:"potato-toml-file" help:"Path to package directory." type:"path" default:"./potato.toml"`
	OutputZipFile  string `name:"output-zip-file" help:"Output path for built package." type:"path"`
}

// operations

type OperationsCmd struct {
	Backup  OperationsBackupCmd  `cmd:"" help:"Backup the database and files."`
	Restore OperationsRestoreCmd `cmd:"" help:"Restore from a backup."`
}

type OperationsBackupCmd struct {
	Output string `name:"output" short:"o" help:"Backup output path." type:"path"`
}

func (c *OperationsBackupCmd) Run(ctx *kong.Context) error {
	panic("Operations Backup - Not implemented yet")

}

type OperationsRestoreCmd struct {
	Input string `arg:"" help:"Backup file to restore from." type:"path"`
	Force bool   `name:"force" short:"f" help:"Force restore without confirmation."`
}

func (c *OperationsRestoreCmd) Run(ctx *kong.Context) error {
	panic("Operations Restore - Not implemented yet")

}

// singleton

type SingletonCmd struct {
	Start SingletonStartCmd `cmd:"" help:"Start the singleton."`
}

type SingletonStartCmd struct {
	Port           int    `name:"port" short:"p" help:"Server port." default:"7777"`
	PackageOutPath string `name:"package-out-path" short:"pop" help:"Package output path." default:"./.single"`
}

// dev

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

// extra

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

func Run() {
	var cli CLI
	parser := kong.Must(&cli,
		kong.Name("potatoverse"),
		kong.Description("Potatoverse: Platform for apps."),
		kong.UsageOnError(),
	)

	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	err = ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
