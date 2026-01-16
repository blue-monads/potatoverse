package repohub

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

// https://github.com/blue-monads/store/raw/refs/heads/master/
// harvest-index.json

var (
	harvestRepos      map[string]*HarvestRepo
	harvestReposMutex sync.Mutex
)

func getHarvestRepo(baseURL string) *HarvestRepo {
	harvestReposMutex.Lock()
	defer harvestReposMutex.Unlock()
	repo, ok := harvestRepos[baseURL]
	if !ok {
		repo = NewHarvestRepo(baseURL)
		harvestRepos[baseURL] = repo
	}
	return repo
}

type PotatoField struct {
	Name               string          `json:"name"`
	Info               string          `json:"info"`
	Type               string          `json:"type"`
	ZipTemplate        string          `json:"zip_template"`
	IndexedTags        []string        `json:"indexed_tags"`
	IndexedTagTemplate string          `json:"indexed_tag_template"`
	Potatoes           []PotatoPackage `json:"potatoes"`
}

type HarvestRepo struct {
	BaseURL    string
	cache      *PotatoField
	cacheTime  time.Time
	cacheMutex sync.Mutex
}

func NewHarvestRepo(baseURL string) *HarvestRepo {
	return &HarvestRepo{
		BaseURL: baseURL,
	}
}

func (h *HarvestRepo) isCacheValid() bool {
	return time.Since(h.cacheTime) < 10*time.Minute
}

func (h *HarvestRepo) getCache() (*PotatoField, error) {
	h.cacheMutex.Lock()
	if h.cache != nil && h.isCacheValid() {
		h.cacheMutex.Unlock()
		return h.cache, nil
	}

	resp, err := http.Get(fmt.Sprintf("%sharvest-index.json", h.BaseURL))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var field PotatoField
	err = json.Unmarshal(body, &field)
	if err != nil {
		return nil, err
	}

	h.cacheMutex.Lock()
	h.cache = &field
	h.cacheTime = time.Now()
	h.cacheMutex.Unlock()

	return &field, nil
}

func (h *HarvestRepo) ListPackages() ([]PotatoPackage, error) {

	field, err := h.getCache()
	if err != nil {
		return nil, err
	}

	return field.Potatoes, nil
}

func (h *HarvestRepo) DownloadPackge(slug string, version string) (string, error) {
	field, err := h.getCache()
	if err != nil {
		return "", err
	}

	potatoIndex := slices.IndexFunc(field.Potatoes, func(p PotatoPackage) bool {
		return p.Slug == slug
	})

	if potatoIndex == -1 {
		return "", fmt.Errorf("package not found: %s", slug)
	}

	potato := &field.Potatoes[potatoIndex]

	if version == "" {
		version = potato.Version
	}

	if version == "" {
		version = potato.Versions[len(potato.Versions)-1]
	}

	url := strings.ReplaceAll(field.ZipTemplate, "{slug}", slug)
	url = strings.ReplaceAll(url, "{version}", version)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	tmpFile, err := os.CreateTemp("", "potato-package-*.zip")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	_, err = tmpFile.Write(body)
	if err != nil {
		return "", err
	}

	return tmpFile.Name(), nil

}
