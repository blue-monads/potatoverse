package cli

import "github.com/alecthomas/kong"

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
