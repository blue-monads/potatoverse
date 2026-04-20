package docs

import (
	"embed"
)

//go:embed all:contents/*
//go:embed all:skills/*
var Docs embed.FS
