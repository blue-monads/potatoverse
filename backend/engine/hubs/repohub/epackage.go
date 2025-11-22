package repohub

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/blue-monads/turnix/backend/xtypes/models"
)

//go:embed all:epackages/*
var embedPackages embed.FS

func ListEPackages() ([]models.PotatoPackage, error) {
	return listEmbeddedPackagesFromFS()
}

// ListEPackagesFromRepo lists packages from a specific repo
func ListEPackagesFromRepo(repoHub *RepoHub, repoSlug string) ([]models.PotatoPackage, error) {
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
func ZipEPackageFromRepo(repoHub *RepoHub, repoSlug string, packageName string) (string, error) {
	if repoHub == nil {
		return zipEmbeddedPackageFromFS(packageName)
	}
	return repoHub.ZipPackage(repoSlug, packageName)
}

func listEmbeddedPackagesFromFS() ([]models.PotatoPackage, error) {
	files, err := embedPackages.ReadDir("epackages")
	if err != nil {
		return nil, err
	}

	epackages := []models.PotatoPackage{}

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

		epackage := models.PotatoPackage{}
		err = json.Unmarshal(jsonFile, &epackage)
		if err != nil {
			// Skip invalid packages
			continue
		}

		epackages = append(epackages, epackage)
	}

	return epackages, nil
}
