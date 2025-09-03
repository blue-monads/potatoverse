package models

import (
	"time"
)

type BprintInstall struct {
	Slug        string     `json:"slug" db:"slug"`
	Type        string     `json:"type" db:"type"`
	StorageType string     `json:"storage_type" db:"storage_type"`
	Reference   string     `json:"reference" db:"reference"`
	Name        string     `json:"name" db:"name"`
	Info        string     `json:"info" db:"info"`
	Tags        string     `json:"tags" db:"tags"`
	InstalledBy int64      `json:"installed_by" db:"installed_by"`
	InstalledAt *time.Time `json:"installed_at" db:"installed_at"`
}

type BprintInstallFile struct {
	ID         int64      `json:"id" db:"id,omitempty"`
	BprintSlug string     `json:"bprint_slug" db:"bprint_slug"`
	Name       string     `json:"name" db:"name"`
	IsFolder   bool       `json:"is_folder" db:"is_folder"`
	Path       string     `json:"path" db:"path"`
	Size       int64      `json:"size" db:"size"`
	Mime       string     `json:"mime" db:"mime"`
	Hash       string     `json:"hash" db:"hash"`
	StoreType  int64      `json:"store_type" db:"store_type"`
	External   bool       `json:"external" db:"external"`
	CreatedBy  int64      `json:"created_by" db:"created_by"`
	CreatedAt  *time.Time `json:"created_at" db:"created_at"`
}

type BprintInstallFileBlob struct {
	ID     int64  `db:"id" json:"id"`
	FileId int64  `db:"file_id" json:"file_id"`
	Size   int64  `db:"size" json:"size"`
	PartId int64  `db:"part_id" json:"part_id"`
	Blob   []byte `db:"blob" json:"blob"`
}

type BprintInstallFileBlobLite struct {
	ID     int64 `db:"id" json:"id"`
	FileId int64 `db:"file_id" json:"file_id"`
	Size   int64 `db:"size" json:"size"`
	PartId int64 `db:"part_id" json:"part_id"`
}
