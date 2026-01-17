package dbmodels

import "time"

type InstalledPackage struct {
	ID              int64      `json:"id" db:"id,omitempty"`
	Name            string     `json:"name" db:"name"`
	InstallRepo     string     `json:"install_repo" db:"install_repo"`
	CanonicalUrl    string     `json:"canonical_url" db:"canonical_url,omitempty"`
	StorageType     string     `json:"storage_type" db:"storage_type"`
	ActiveInstallID int64      `json:"active_install_id" db:"active_install_id"`
	InstalledBy     int64      `json:"installed_by" db:"installed_by"`
	InstalledAt     *time.Time `json:"installed_at" db:"installed_at,omitempty"`
	DevToken        string     `json:"dev_token" db:"dev_token"`
}

type PackageVersion struct {
	ID            int64  `json:"id" db:"id,omitempty"`
	InstallId     int64  `json:"install_id" db:"install_id"`
	Name          string `json:"name" db:"name"`
	Slug          string `json:"slug" db:"slug"`
	Info          string `json:"info" db:"info,omitempty"`
	Tags          string `json:"tags" db:"tags"`
	FormatVersion string `json:"format_version" db:"format_version"`
	AuthorName    string `json:"author_name" db:"author_name"`
	AuthorEmail   string `json:"author_email" db:"author_email"`
	AuthorSite    string `json:"author_site" db:"author_site"`
	SourceCode    string `json:"source_code" db:"source_code"`
	License       string `json:"license" db:"license"`
	Version       string `json:"version" db:"version"`
	InitPage      string `json:"init_page" db:"init_page,omitempty"`
	UpdatePage    string `json:"update_page" db:"update_page,omitempty"`
}
