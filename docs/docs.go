package docs

import (
	"embed"
)

//go:embed all:contents/*
var Docs embed.FS
