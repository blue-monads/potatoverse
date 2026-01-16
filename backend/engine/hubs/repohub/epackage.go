package repohub

import (
	"archive/zip"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
)

//go:embed all:epackages/*
var embedPackages embed.FS

func ListEPackages() ([]PotatoPackage, error) {
	return listEmbeddedPackagesFromFS()
}

// ListEPackagesFromRepo lists packages from a specific repo
func ListEPackagesFromRepo(repoHub *RepoHub, repoSlug string) ([]PotatoPackage, error) {
	if repoHub == nil {
		return listEmbeddedPackagesFromFS()
	}
	return repoHub.ListPackages(repoSlug)
}

// ZipEPackage creates a zip from embedded package (for backward compatibility)
func ZipEPackage(name string) (string, error) {
	return zipEmbeddedPackageFromFS(name)
}

// ZipEPackageFromRepo creates a zip from a package in a specific repo
func ZipEPackageFromRepo(repoHub *RepoHub, repoSlug string, packageName string, version string) (string, error) {
	if repoHub == nil {
		return zipEmbeddedPackageFromFS(packageName)
	}
	return repoHub.ZipPackage(repoSlug, packageName, version)
}

func listEmbeddedPackagesFromFS() ([]PotatoPackage, error) {
	files, err := embedPackages.ReadDir("epackages")
	if err != nil {
		return nil, err
	}

	epackages := []PotatoPackage{}

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

		epackage := PotatoPackage{}
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
	zipFile, err := os.CreateTemp("", "turnix-package-*.zip")
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
