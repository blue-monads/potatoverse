package devrepo

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/providers/erepo"
	"github.com/blue-monads/potatoverse/backend/xtypes/models"
)

func TestDevRepo_PackagesDiscovery(t *testing.T) {
	repo := erepo.NewEmbedRepo(embedPackages)

	pkgs, err := repo.ListPackages()
	if err != nil {
		t.Fatalf("failed to list packages: %v", err)
	}

	foundAlpha := false
	foundBeta := false

	for _, pkg := range pkgs {
		if pkg.Slug == "pexample-alpha" {
			foundAlpha = true
		}
		if pkg.Slug == "pexample-beta" {
			foundBeta = true
		}
	}

	if !foundAlpha {
		t.Errorf("pexample-alpha was not found in devrepo")
	}
	if !foundBeta {
		t.Errorf("pexample-beta was not found in devrepo")
	}

	// Verify pexample-alpha manifest
	alphaRaw, err := embedPackages.ReadFile("epackages/pexample-alpha/potato.json")
	if err != nil {
		t.Fatalf("failed to read pexample-alpha potato.json: %v", err)
	}
	var alphaManifest models.PotatoPackage
	if err := json.Unmarshal(alphaRaw, &alphaManifest); err != nil {
		t.Fatalf("failed to unmarshal pexample-alpha manifest: %v", err)
	}
	if len(alphaManifest.Spaces) == 0 {
		t.Errorf("expected spaces in pexample-alpha")
	} else if alphaManifest.Spaces[0].Namespace != "pexample-alpha" {
		t.Errorf("expected space namespace pexample-alpha, got %s", alphaManifest.Spaces[0].Namespace)
	}

	// Verify pexample-beta manifest
	betaRaw, err := embedPackages.ReadFile("epackages/pexample-beta/potato.json")
	if err != nil {
		t.Fatalf("failed to read pexample-beta potato.json: %v", err)
	}
	var betaManifest models.PotatoPackage
	if err := json.Unmarshal(betaRaw, &betaManifest); err != nil {
		t.Fatalf("failed to unmarshal pexample-beta manifest: %v", err)
	}
	if len(betaManifest.Spaces) == 0 {
		t.Errorf("expected spaces in pexample-beta")
	} else {
		space := betaManifest.Spaces[0]
		if space.Namespace != "pexample-beta" {
			t.Errorf("expected space namespace pexample-beta, got %s", space.Namespace)
		}
		if space.SpaceType != "AppPlugin" {
			t.Errorf("expected space_type AppPlugin, got %s", space.SpaceType)
		}
		if space.LoaderScript != "loader.js" {
			t.Errorf("expected loader_script loader.js, got %s", space.LoaderScript)
		}
	}

	// Verify loader.js exists
	loaderJs, err := embedPackages.ReadFile("epackages/pexample-beta/loader.js")
	if err != nil || len(loaderJs) == 0 {
		t.Fatalf("expected non-empty loader.js in pexample-beta: %v", err)
	}
}

func TestDevRepo_ZipPackages(t *testing.T) {
	repo := erepo.NewEmbedRepo(embedPackages)

	zipAlpha, err := repo.ZipPackage("pexample-alpha", "0.0.1")
	if err != nil {
		t.Fatalf("failed to zip pexample-alpha: %v", err)
	}
	defer os.Remove(zipAlpha)
	if zipAlpha == "" {
		t.Errorf("expected zip file path, got empty")
	}

	zipBeta, err := repo.ZipPackage("pexample-beta", "0.0.1")
	if err != nil {
		t.Fatalf("failed to zip pexample-beta: %v", err)
	}
	defer os.Remove(zipBeta)
	if zipBeta == "" {
		t.Errorf("expected zip file path, got empty")
	}
}
