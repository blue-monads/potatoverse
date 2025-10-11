package dbmodels

import "time"

type File struct {
	ID           int64      `db:"id,omitempty" json:"id"`
	Name         string     `db:"name" json:"name"`
	Path         string     `db:"path" json:"path"`
	Size         int64      `db:"size" json:"size"`
	Mime         string     `db:"mime" json:"mime"`
	Hash         string     `db:"hash" json:"hash"`
	IsFolder     bool       `db:"is_folder" json:"is_folder"`
	StoreType    int64      `db:"storeType" json:"storeType"`
	OwnerSpaceID int64      `db:"owner_space_id" json:"owner_space_id"`
	CreatedBy    int64      `db:"created_by" json:"created_by"`
	CreatedAt    *time.Time `db:"created_at" json:"created_at"`
}

type FilePartedBlob struct {
	Id     int64 `db:"id" json:"id"`
	FileId int64 `db:"file_id" json:"file_id"`
	Size   int64 `db:"size" json:"size"`
	PartId int64 `db:"part_id" json:"part_id"`
}

type FileShare struct {
	ID        string     `json:"id" db:"id,omitempty"`
	FileID    int64      `json:"file_id" db:"file_id"`
	UserID    int64      `json:"user_id" db:"user_id"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
}
