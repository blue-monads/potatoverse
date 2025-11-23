package repohub

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/models"
)

type RepoHub struct {
	repos    map[string]*xtypes.RepoOptions
	logger   *slog.Logger
	httpPort int
}

func NewRepoHub(repos []xtypes.RepoOptions, logger *slog.Logger, port int) *RepoHub {
	hub := &RepoHub{
		repos:    make(map[string]*xtypes.RepoOptions),
		logger:   logger,
		httpPort: port,
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
	body, err := h.getPackageIndexFromHttp(url)
	if err != nil {
		return nil, err
	}

	var packages []models.PotatoPackage
	err = json.Unmarshal(body, &packages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package list: %w", err)
	}

	return packages, nil
}

func (h *RepoHub) getPackageIndexFromHttp(url string) ([]byte, error) {
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

	return body, nil

}

// zipEmbeddedPackage creates a zip file from embedded package
func (h *RepoHub) zipEmbeddedPackage(name string) (string, error) {
	return zipEmbeddedPackageFromFS(name)
}

// zipHttpPackage downloads and creates a zip file from HTTP package
func (h *RepoHub) zipHttpPackage(baseURL string, packageName string) (string, error) {

	pIndex, err := h.getPackageIndexFromHttp(baseURL)
	if err != nil {
		return "", err
	}

	var packages HttpPackageIndexV1
	err = json.Unmarshal(pIndex, &packages)
	if err != nil {
		return "", fmt.Errorf("failed to parse package list: %w", err)
	}

	fileUrl := ""

	for i := range packages {
		pkg := &packages[i]
		if pkg.Slug == packageName || pkg.Name == packageName {
			fileUrl = pkg.DownloadUrl
			break
		}
	}

	if strings.HasPrefix(fileUrl, "/") {
		fileUrl = fmt.Sprintf("http://localhost:%d%s", h.httpPort, fileUrl)
	}

	if fileUrl == "" {
		return "", fmt.Errorf("package not found: %s", packageName)
	}

	resp, err := http.Get(fileUrl)
	if err != nil {
		return "", fmt.Errorf("failed to download package: %w", err)
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
