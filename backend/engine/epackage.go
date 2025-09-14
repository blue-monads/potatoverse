package engine

import "embed"

//go:embed all:packages/*
var embedPackages embed.FS
