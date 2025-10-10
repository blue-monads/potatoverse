package engine

import (
	"archive/zip"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/blue-monads/turnix/backend/xtypes/models"
	"github.com/k0kubun/pp"
)

//go:embed all:epackages/*
var embedPackages embed.FS

func ListEPackages() ([]models.PotatoPackage, error) {
	files, err := embedPackages.ReadDir("epackages")
	if err != nil {
		return nil, err
	}

	epackages := []models.PotatoPackage{}

	for _, file := range files {

		pp.Println("@file", file.Name())

		if !file.IsDir() {
			continue
		}

		fileName := fmt.Sprintf("epackages/%s/potato.json", file.Name())

		jsonFile, err := embedPackages.ReadFile(fileName)
		if err != nil {
			return nil, err
		}

		epackage := models.PotatoPackage{}
		err = json.Unmarshal(jsonFile, &epackage)
		if err != nil {
			return nil, err
		}

		epackages = append(epackages, epackage)

	}

	return epackages, nil
}

func ZipEPackage(name string) (string, error) {

	zipFile, err := os.CreateTemp("", "turnix-package-*.zip")
	if err != nil {
		return "", err
	}

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = includeSubFolder(name, "", zipWriter)

	return zipFile.Name(), nil
}

func includeSubFolder(name, folder string, zipWriter *zip.Writer) error {
	readPath := path.Join("epackages/", name, folder)

	files, err := embedPackages.ReadDir(readPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		targetPath := path.Join(folder, file.Name())
		targetPath = strings.TrimLeft(targetPath, "/")

		if file.IsDir() {
			err = includeSubFolder(name, targetPath, zipWriter)
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
