package cli

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/joho/godotenv"
)

type CLI struct {
	Server     ServerCmd     `cmd:"" help:"Server management commands."`
	Package    PackageCmd    `cmd:"" help:"Package management commands."`
	Operations OperationsCmd `cmd:"" help:"Backup and restore operations."`
	Dev        DevCmd        `cmd:"" help:"Development utilities."`
	Extra      ExtraCmd      `cmd:"" help:"Extra commands."`
	Verbose    bool          `name:"verbose" short:"v" help:"Enable verbose output."`
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

// extra

func loadEnv() {

	godotenv.Load(".env.potato")

}

func Run() {
	var cli CLI
	parser := kong.Must(&cli,
		kong.Name("potatoverse"),
		kong.Description("Potatoverse: Platform for apps."),
		kong.UsageOnError(),
	)

	loadEnv()

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
