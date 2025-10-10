package models

import "time"

type Package struct {
	ID          int64      `json:"id" db:"id,omitempty"`
	Name        string     `json:"name" db:"name"`
	Slug        string     `json:"slug" db:"slug"`
	Info        string     `json:"info" db:"info,omitempty"`
	Tags        string     `json:"tags" db:"tags"`
	StorageType string     `json:"storage_type" db:"storage_type"`
	Reference   string     `json:"reference" db:"reference"`
	InstalledBy int64      `json:"installed_by" db:"installed_by"`
	InstalledAt *time.Time `json:"installed_at" db:"installed_at,omitempty"`
}

type PackageFile struct {
	ID        int64      `json:"id" db:"id,omitempty"`
	PackageID int64      `json:"package_id" db:"package_id"`
	Name      string     `json:"name" db:"name"`
	IsFolder  bool       `json:"is_folder" db:"is_folder"`
	Path      string     `json:"path" db:"path"`
	Size      int64      `json:"size" db:"size"`
	Mime      string     `json:"mime" db:"mime"`
	Hash      string     `json:"hash" db:"hash"`
	StoreType int64      `json:"store_type" db:"store_type"`
	CreatedBy int64      `json:"created_by" db:"created_by"`
	CreatedAt *time.Time `json:"created_at" db:"created_at,omitempty"`
}

type PackageFileBlob struct {
	ID     int64  `json:"id" db:"id,omitempty"`
	FileID int64  `json:"file_id" db:"file_id"`
	Size   int64  `json:"size" db:"size"`
	PartID int64  `json:"part_id" db:"part_id"`
	Blob   []byte `json:"blob" db:"blob"`
}

type PackageFileBlobLite struct {
	ID     int64 `json:"id" db:"id,omitempty"`
	FileID int64 `json:"file_id" db:"file_id"`
	Size   int64 `json:"size" db:"size"`
	PartID int64 `json:"part_id" db:"part_id"`
}
