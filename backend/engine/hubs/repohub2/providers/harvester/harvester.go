package harvester

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

	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub2/repotypes"
)

var (
	_ repotypes.IRepo = (*HarvesterRepo)(nil)
)

type HarvesterRepo struct {
	baseURL    string
	cache      *PotatoField
	cacheTime  time.Time
	cacheMutex sync.Mutex
}

func NewHarvesterRepo(baseURL string) *HarvesterRepo {
	return &HarvesterRepo{baseURL: baseURL}
}

func (r *HarvesterRepo) isCacheValid() bool {
	return time.Since(r.cacheTime) < 10*time.Minute
}

func (r *HarvesterRepo) getCache() (*PotatoField, error) {
	r.cacheMutex.Lock()
	if r.cache != nil && r.isCacheValid() {
		r.cacheMutex.Unlock()
		return r.cache, nil
	}

	resp, err := http.Get(fmt.Sprintf("%sharvest-index.json", r.baseURL))
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

	r.cache = &field
	r.cacheTime = time.Now()
	r.cacheMutex.Unlock()

	return &field, nil
}

func (r *HarvesterRepo) ListPackages() ([]repotypes.PotatoPackage, error) {
	field, err := r.getCache()
	if err != nil {
		return nil, err
	}

	return field.Potatoes, nil
}

func (r *HarvesterRepo) ZipPackage(packageName string, version string) (string, error) {
	field, err := r.getCache()
	if err != nil {
		return "", err
	}

	potatoIndex := slices.IndexFunc(field.Potatoes, func(p repotypes.PotatoPackage) bool {
		return p.Slug == packageName
	})

	if potatoIndex == -1 {
		return "", fmt.Errorf("package not found: %s", packageName)
	}

	potato := &field.Potatoes[potatoIndex]

	if version == "" {
		version = potato.Version
	}

	if version == "" {
		version = potato.Versions[len(potato.Versions)-1]
	}

	url := strings.ReplaceAll(field.ZipTemplate, "{slug}", packageName)
	url = strings.ReplaceAll(url, "{version}", version)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	tmpFile, err := os.CreateTemp("", "potato-package-*.zip")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}
