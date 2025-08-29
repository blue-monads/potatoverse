package frontend

import "embed"

//go:embed all:output/*
var BuildProd embed.FS
