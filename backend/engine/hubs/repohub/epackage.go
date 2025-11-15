package repohub

import (
	"embed"

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
