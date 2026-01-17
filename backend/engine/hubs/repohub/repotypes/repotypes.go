package repotypes

import (
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

type IRepoHub interface {
	Run(app xtypes.App) error
	ListRepos() []xtypes.RepoOptions
	GetRepo(slug string) (*xtypes.RepoOptions, error)
	ListPackages(repoSlug string) ([]PotatoPackage, error)
	ZipPackage(repoSlug string, packageName string, version string) (string, error)
}

type RepoProvider func(app xtypes.App, repoOptions *xtypes.RepoOptions) (IRepo, error)

type IRepo interface {
	ListPackages() ([]PotatoPackage, error)
	ZipPackage(packageName string, version string) (string, error)
}

type PotatoPackage struct {
	Name          string   `json:"name" toml:"name"`
	Slug          string   `json:"slug" toml:"slug"`
	Info          string   `json:"info" toml:"info"`
	Tags          []string `json:"tags" toml:"tags"`
	FormatVersion string   `json:"format_version" toml:"format_version"`
	AuthorName    string   `json:"author_name" toml:"author_name"`
	AuthorEmail   string   `json:"author_email" toml:"author_email"`
	AuthorSite    string   `json:"author_site" toml:"author_site"`
	SourceCode    string   `json:"source_code" toml:"source_code"`
	License       string   `json:"license" toml:"license"`
	Version       string   `json:"version" toml:"version"`
	Versions      []string `json:"versions" toml:"versions"`
}
