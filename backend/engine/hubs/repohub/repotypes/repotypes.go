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
	Name          string   `json:"name" yaml:"name"`
	Slug          string   `json:"slug" yaml:"slug"`
	Info          string   `json:"info" yaml:"info"`
	Tags          []string `json:"tags" yaml:"tags"`
	FormatVersion string   `json:"format_version" yaml:"format_version"`
	AuthorName    string   `json:"author_name" yaml:"author_name"`
	AuthorEmail   string   `json:"author_email" yaml:"author_email"`
	AuthorSite    string   `json:"author_site" yaml:"author_site"`
	SourceCode    string   `json:"source_code" yaml:"source_code"`
	License       string   `json:"license" yaml:"license"`
	Version       string   `json:"version" yaml:"version"`
	Versions      []string `json:"versions" yaml:"versions"`
}
