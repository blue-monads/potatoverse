package repohub

import (
	"fmt"
	"maps"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub2/repotypes"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

type RepoHub struct {
	repos   map[string]repotypes.IRepo
	options []xtypes.RepoOptions
}

func NewRepoHub(repos []xtypes.RepoOptions) *RepoHub {
	return &RepoHub{
		repos:   make(map[string]repotypes.IRepo),
		options: repos,
	}
}

func (h *RepoHub) Run(app xtypes.App) error {

	repoProvidersMutex.RLock()
	providers := maps.Clone(repoProviders)
	repoProvidersMutex.RUnlock()

	for _, option := range h.options {

		provider, ok := providers[option.Type]
		if !ok {
			return fmt.Errorf("repo provider not found: %s", option.Type)
		}

		repo, err := provider(app, &option)
		if err != nil {
			return err
		}
		h.repos[option.Slug] = repo
	}
	return nil
}

func (h *RepoHub) ListRepos() []xtypes.RepoOptions {
	return h.options
}

func (h *RepoHub) GetRepo(slug string) (*xtypes.RepoOptions, error) {
	for _, option := range h.options {
		if option.Slug == slug {
			return &option, nil
		}
	}

	return nil, fmt.Errorf("repo not found: %s", slug)
}

func (h *RepoHub) ListPackages(repoSlug string) ([]repotypes.PotatoPackage, error) {
	repo := h.repos[repoSlug]
	if repo == nil {
		return nil, fmt.Errorf("repo not found: %s", repoSlug)
	}

	return repo.ListPackages()
}

func (h *RepoHub) ZipPackage(repoSlug string, packageName string, version string) (string, error) {
	repo := h.repos[repoSlug]
	if repo == nil {
		return "", fmt.Errorf("repo not found: %s", repoSlug)
	}

	return repo.ZipPackage(packageName, version)
}
