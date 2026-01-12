package cli

import "github.com/alecthomas/kong"

type FieldCmd struct {
	Init FieldInitCmd `cmd:"" help:"Initialize the field."`
}

type FieldInitCmd struct{}

func (c *FieldInitCmd) Run(ctx *kong.Context) error {
	return nil
}
