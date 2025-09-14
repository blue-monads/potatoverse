package engine

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/k0kubun/pp"
)

//go:embed all:epackages/*
var embedPackages embed.FS

type EPackage struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
	Type        string `json:"type"`
	Tags        string `json:"tags"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	TimeAgo     string `json:"timeAgo"`
	MCp         bool   `json:"mcp"`
}

func ListEPackages() ([]EPackage, error) {
	files, err := embedPackages.ReadDir("epackages")
	if err != nil {
		return nil, err
	}

	epackages := []EPackage{}

	for _, file := range files {

		pp.Println("@file", file.Name())

		if !file.IsDir() {
			continue
		}

		fileName := fmt.Sprintf("epackages/%s/turnix.json", file.Name())

		jsonFile, err := embedPackages.ReadFile(fileName)
		if err != nil {
			return nil, err
		}

		epackage := EPackage{}
		err = json.Unmarshal(jsonFile, &epackage)
		if err != nil {
			return nil, err
		}

		epackage.Author = "Demo"
		epackage.TimeAgo = "Just now"
		epackage.MCp = false

		epackages = append(epackages, epackage)

	}

	return epackages, nil
}
