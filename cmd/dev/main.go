package main

import (
	_ "github.com/mattn/go-sqlite3"

	// modules

	"github.com/blue-monads/turnix/v2/backend"

	_ "github.com/blue-monads/turnix/backend/modules/books"
	_ "github.com/blue-monads/turnix/backend/modules/zblog"

	// _ "github.com/blue-monads/turnix/backend/modules/tracker"
	_ "github.com/blue-monads/turnix/backend/modules/unloop"
)

// //go:embed all:assets/*
// var devModeAssets embed.FS

func main() {

	// registry.Register("devmode", func(opt xproject.BuilderOption) (*xproject.Defination, error) {
	// 	return &xproject.Defination{
	// 		Name:            "Devmode",
	// 		Slug:            "devmode",
	// 		Info:            "Development mode",
	// 		Version:         "0.0.1",
	// 		LinkPattern:     "/z/x/devmode",
	// 		AssetData:       devModeAssets,
	// 		AssetDataPrefix: "assets",
	// 	}, nil
	// })

	// devmode.Run()

	backend.NewMain()

}
