package docs

import (
	"embed"
)

//go:embed all:contents/*

var Docs embed.FS

//go:embed all:skills/*

var Skills embed.FS
