package dbmodels

import "time"

type FileMeta struct {
	ID        int64      `db:"id,omitempty" json:"id"`
	RefID     string     `db:"ref_id" json:"ref_id"`
	OwnerID   int64      `db:"owner_id" json:"owner_id"`
	Name      string     `db:"name" json:"name"`
	IsFolder  bool       `db:"is_folder" json:"is_folder"`
	Path      string     `db:"path" json:"path"`
	Size      int64      `db:"size" json:"size"`
	Mime      string     `db:"mime" json:"mime"`
	Hash      string     `db:"hash" json:"hash"`
	StoreType int64      `db:"store_type" json:"store_type"`
	CreatedBy int64      `db:"created_by" json:"created_by"`
	CreatedAt *time.Time `db:"created_at,omitempty" json:"created_at"`
	UpdatedAt *time.Time `db:"updated_at,omitempty" json:"updated_at"`
	UpdatedBy int64      `db:"updated_by" json:"updated_by"`
}

type FileBlob struct {
	Id     int64  `db:"id,omitempty" json:"id"`
	FileID int64  `db:"file_id" json:"file_id"`
	Size   int64  `db:"size" json:"size"`
	PartID int64  `db:"part_id" json:"part_id"`
	Blob   []byte `db:"blob" json:"blob"`
}
