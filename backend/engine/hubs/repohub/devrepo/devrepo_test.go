package devrepo

import (
	"archive/zip"
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
	foundRoot := false

	for _, pkg := range pkgs {
		if pkg.Slug == "pexample-alpha" {
			foundAlpha = true
		}
		if pkg.Slug == "pexample-beta" {
			foundBeta = true
		}
		if pkg.Slug == "pexample-root" {
			foundRoot = true
		}
	}

	if !foundAlpha {
		t.Errorf("pexample-alpha was not found in devrepo")
	}
	if !foundBeta {
		t.Errorf("pexample-beta was not found in devrepo")
	}
	if !foundRoot {
		t.Errorf("pexample-root was not found in devrepo")
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

	// Verify pexample-root manifest (multi-space with root space)
	rootRaw, err := embedPackages.ReadFile("epackages/pexample-root/potato.json")
	if err != nil {
		t.Fatalf("failed to read pexample-root potato.json: %v", err)
	}
	var rootManifest models.PotatoPackage
	if err := json.Unmarshal(rootRaw, &rootManifest); err != nil {
		t.Fatalf("failed to unmarshal pexample-root manifest: %v", err)
	}
	if len(rootManifest.Spaces) != 2 {
		t.Fatalf("expected 2 spaces in pexample-root, got %d", len(rootManifest.Spaces))
	}

	// Space 0: Root Space
	rootSpace := rootManifest.Spaces[0]
	if rootSpace.Namespace != "pexample-root" {
		t.Errorf("expected root space namespace pexample-root, got %s", rootSpace.Namespace)
	}
	if rootSpace.SpaceType != "RootApp" {
		t.Errorf("expected space_type RootApp, got %s", rootSpace.SpaceType)
	}
	if !rootSpace.IsDefault {
		t.Errorf("expected root space is_default true")
	}
	if rootSpace.ServerFile != "server.lua" {
		t.Errorf("expected server.lua, got %s", rootSpace.ServerFile)
	}
	if rootSpace.RouteOptions.ServeFolder != "public" {
		t.Errorf("expected serve_folder public, got %s", rootSpace.RouteOptions.ServeFolder)
	}

	// Space 1: Sub-space
	subSpace := rootManifest.Spaces[1]
	if subSpace.Namespace != "pexample-root:sub" {
		t.Errorf("expected sub-space namespace pexample-root:sub, got %s", subSpace.Namespace)
	}
	if subSpace.SpaceType != "App" {
		t.Errorf("expected sub-space space_type App, got %s", subSpace.SpaceType)
	}
	if subSpace.IsDefault {
		t.Errorf("expected sub-space is_default false")
	}
	if subSpace.ServerFile != "sub/server.lua" {
		t.Errorf("expected sub/server.lua, got %s", subSpace.ServerFile)
	}
	if subSpace.RouteOptions.ServeFolder != "sub/public" {
		t.Errorf("expected serve_folder sub/public, got %s", subSpace.RouteOptions.ServeFolder)
	}

	// Verify pexample-root files exist in embedPackages
	for _, expectedFile := range []string{
		"epackages/pexample-root/server.lua",
		"epackages/pexample-root/public/index.html",
		"epackages/pexample-root/sub/server.lua",
		"epackages/pexample-root/sub/public/index.html",
	} {
		data, err := embedPackages.ReadFile(expectedFile)
		if err != nil || len(data) == 0 {
			t.Errorf("expected non-empty file %s: %v", expectedFile, err)
		}
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

	zipRoot, err := repo.ZipPackage("pexample-root", "0.0.1")
	if err != nil {
		t.Fatalf("failed to zip pexample-root: %v", err)
	}
	defer os.Remove(zipRoot)
	if zipRoot == "" {
		t.Errorf("expected zip file path, got empty")
	}

	// Verify zip contents for pexample-root
	zr, err := zip.OpenReader(zipRoot)
	if err != nil {
		t.Fatalf("failed to open pexample-root zip: %v", err)
	}
	defer zr.Close()

	expectedZipEntries := map[string]bool{
		"potato.json":            false,
		"server.lua":             false,
		"public/index.html":      false,
		"sub/server.lua":         false,
		"sub/public/index.html":  false,
	}
	for _, f := range zr.File {
		if _, ok := expectedZipEntries[f.Name]; ok {
			expectedZipEntries[f.Name] = true
		}
	}
	for entry, found := range expectedZipEntries {
		if !found {
			t.Errorf("expected entry %s in pexample-root zip", entry)
		}
	}
}
