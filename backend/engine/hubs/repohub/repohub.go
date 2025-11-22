package repohub

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/models"
)

type RepoHub struct {
	repos  map[string]*xtypes.RepoOptions
	logger *slog.Logger
}

func NewRepoHub(repos []xtypes.RepoOptions, logger *slog.Logger, port int) *RepoHub {
	hub := &RepoHub{
		repos:  make(map[string]*xtypes.RepoOptions),
		logger: logger,
	}

	for i := range repos {
		repo := &repos[i]
		// Use slug as key, fallback to name if slug is empty
		key := repo.Slug
		if key == "" {
			key = repo.Name
		}
		if key == "" {
			panic("repo slug or name is required")
		}

		if repo.Type == "" {
			panic("repo type is required")
		}

		if repo.Type == "http" && repo.URL == "" {
			panic("http repo url is required")
		}

		if repo.Type == "http" && !strings.HasPrefix(repo.URL, "http") {
			repo.URL = fmt.Sprintf("http://localhost:%d%s", port, repo.URL)
		}

		hub.repos[key] = repo

	}

	return hub
}

// ListRepos returns all available repos
func (h *RepoHub) ListRepos() []xtypes.RepoOptions {

	result := make([]xtypes.RepoOptions, 0, len(h.repos))
	for _, repo := range h.repos {
		result = append(result, *repo)
	}

	return result
}

// GetRepo returns a repo by slug
func (h *RepoHub) GetRepo(slug string) (*xtypes.RepoOptions, error) {
	repo, ok := h.repos[slug]
	if !ok {
		return nil, fmt.Errorf("repo not found: %s", slug)
	}
	return repo, nil
}

// ListPackages lists packages from a specific repo
func (h *RepoHub) ListPackages(repoSlug string) ([]models.PotatoPackage, error) {
	repo, err := h.GetRepo(repoSlug)
	if err != nil {
		return nil, err
	}

	switch repo.Type {
	case "embed", "embeded", "embedded":
		return h.listEmbeddedPackages()
	case "http":
		return h.listHttpPackages(repo.URL)
	default:
		return nil, fmt.Errorf("unsupported repo type: %s", repo.Type)
	}
}

// ZipPackage creates a zip file for a package from a specific repo
func (h *RepoHub) ZipPackage(repoSlug string, packageName string) (string, error) {
	repo, err := h.GetRepo(repoSlug)
	if err != nil {
		return "", err
	}

	switch repo.Type {
	case "embed", "embeded", "embedded":
		return h.zipEmbeddedPackage(packageName)
	case "http":
		return h.zipHttpPackage(repo.URL, packageName)
	default:
		return "", fmt.Errorf("unsupported repo type: %s", repo.Type)
	}
}

// listEmbeddedPackages lists packages from embedded filesystem
func (h *RepoHub) listEmbeddedPackages() ([]models.PotatoPackage, error) {
	return listEmbeddedPackagesFromFS()
}

// listHttpPackages fetches package list from HTTP endpoint
func (h *RepoHub) listHttpPackages(url string) ([]models.PotatoPackage, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch package list: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var packages []models.PotatoPackage
	err = json.Unmarshal(body, &packages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package list: %w", err)
	}

	return packages, nil
}

// zipEmbeddedPackage creates a zip file from embedded package
func (h *RepoHub) zipEmbeddedPackage(name string) (string, error) {
	return zipEmbeddedPackageFromFS(name)
}

// zipHttpPackage downloads and creates a zip file from HTTP package
func (h *RepoHub) zipHttpPackage(baseURL string, packageName string) (string, error) {
	// Construct the package download URL
	// This assumes the package URL follows a pattern like: baseURL/packages/{name}.zip
	// or baseURL/{name}.zip
	// We'll try both patterns
	urls := []string{
		fmt.Sprintf("%s/packages/%s.zip", strings.TrimSuffix(baseURL, "/"), packageName),
		fmt.Sprintf("%s/%s.zip", strings.TrimSuffix(baseURL, "/"), packageName),
	}

	var resp *http.Response
	var err error
	for _, url := range urls {
		resp, err = http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		// If direct download doesn't work, try to get package info first
		// and use the UpdateUrl if available
		packages, err := h.listHttpPackages(baseURL)
		if err != nil {
			return "", fmt.Errorf("failed to get package info: %w", err)
		}

		var pkg *models.PotatoPackage
		for i := range packages {
			if packages[i].Slug == packageName || packages[i].Name == packageName {
				pkg = &packages[i]
				break
			}
		}

		if pkg == nil {
			return "", fmt.Errorf("package not found: %s", packageName)
		}

		resp, err = http.Get(pkg.CanonicalUrl)
		if err != nil {
			return "", fmt.Errorf("failed to download package: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to download package: status %d", resp.StatusCode)
		}
	}
	defer resp.Body.Close()

	tmpFile, err := os.CreateTemp("", "turnix-package-*.zip")
	if err != nil {
		return "", err
	}

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to save package: %w", err)
	}
	tmpFile.Close()

	return tmpFile.Name(), nil
}

// zipEmbeddedPackageFromFS creates a zip file from embedded package (internal helper)
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
