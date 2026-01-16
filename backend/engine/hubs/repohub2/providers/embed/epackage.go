package embed

import (
	"archive/zip"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub2/repotypes"
)

//go:embed all:epackages/*
var embedPackages embed.FS

func ListEPackages() ([]repotypes.PotatoPackage, error) {
	return listEmbeddedPackagesFromFS()
}

// ZipEPackage creates a zip from embedded package (for backward compatibility)
func ZipEPackage(name string) (string, error) {
	return zipEmbeddedPackageFromFS(name)
}

func listEmbeddedPackagesFromFS() ([]repotypes.PotatoPackage, error) {
	files, err := embedPackages.ReadDir("epackages")
	if err != nil {
		return nil, err
	}

	epackages := []repotypes.PotatoPackage{}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		fileName := fmt.Sprintf("epackages/%s/potato.json", file.Name())

		jsonFile, err := embedPackages.ReadFile(fileName)
		if err != nil {
			// Skip if potato.json doesn't exist
			continue
		}

		epackage := repotypes.PotatoPackage{}
		err = json.Unmarshal(jsonFile, &epackage)
		if err != nil {
			// Skip invalid packages
			continue
		}

		epackages = append(epackages, epackage)
	}

	return epackages, nil
}

func zipEmbeddedPackageFromFS(name string) (string, error) {
	zipFile, err := os.CreateTemp("", "potato-package-*.zip")
	if err != nil {
		return "", err
	}

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = includeSubFolderFromFS(name, "", zipWriter)
	if err != nil {
		return "", err
	}

	return zipFile.Name(), nil
}

// includeSubFolderFromFS recursively includes files from embedded filesystem
func includeSubFolderFromFS(name, folder string, zipWriter *zip.Writer) error {
	readPath := path.Join("epackages/", name, folder)

	files, err := embedPackages.ReadDir(readPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		targetPath := path.Join(folder, file.Name())
		targetPath = strings.TrimLeft(targetPath, "/")

		if file.IsDir() {
			err = includeSubFolderFromFS(name, targetPath, zipWriter)
			if err != nil {
				return err
			}
			continue
		}
		fileWriter, err := zipWriter.Create(targetPath)
		if err != nil {
			return err
		}

		finalpath := path.Join(readPath, file.Name())

		fileData, err := embedPackages.ReadFile(finalpath)
		if err != nil {
			return err
		}
		_, err = fileWriter.Write(fileData)
		if err != nil {
			return err
		}
	}
	return nil
}
