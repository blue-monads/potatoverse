package file

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/upper/db/v4"
)

const BlobSizeLimit = 1024 * 1024 * 1

const (
	StoreTypeInline    = 0
	StoreTypeExternal  = 1
	StoreTypeMultipart = 2
)

type FileContext struct {
	Type      string
	OwnerID   int64
	CreatedBy int64
}

type FileMeta struct {
	ID        int64      `db:"id" json:"id,omitempty"`
	OwnerID   int64      `db:"owner_id" json:"owner_id,omitempty"`
	Name      string     `db:"name" json:"name,omitempty"`
	IsFolder  bool       `db:"is_folder" json:"is_folder,omitempty"`
	Path      string     `db:"path" json:"path,omitempty"`
	Size      int64      `db:"size" json:"size,omitempty"`
	Mime      string     `db:"mime" json:"mime,omitempty"`
	Hash      string     `db:"hash" json:"hash,omitempty"`
	StoreType int64      `db:"store_type" json:"store_type,omitempty"`
	CreatedAt *time.Time `db:"created_at" json:"created_at,omitempty"`
	CreatedBy int64      `db:"created_by" json:"created_by,omitempty"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
	UpdatedBy int64      `db:"updated_by" json:"updated_by,omitempty"`
}

type FileBlob struct {
	ID     int64  `db:"id"`
	FileID int64  `db:"file_id"`
	Size   int64  `db:"size"`
	PartID int64  `db:"part_id"`
	Blob   []byte `db:"blob"`
}

type FileBlobLite struct {
	ID     int64 `db:"id"`
	FileID int64 `db:"file_id"`
	Size   int64 `db:"size"`
	PartID int64 `db:"part_id"`
}

type CreateFileRequest struct {
	OwnerID   int64  `json:"owner_id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	StoreType int64  `json:"store_type"`
	CreatedBy int64  `json:"created_by"`
	IsFolder  bool   `json:"is_folder"`
}

type FileOperations struct {
	db                db.Session
	minMultiPartSize  int64
	externalFileMode  bool
	externalFilesPath string
	prefix            string
}

type Options struct {
	DbSess            db.Session
	MinMultiPartSize  int64
	ExternalFileMode  bool
	ExternalFilesPath string
	Prefix            string
}

func NewFileOperations(opts Options) *FileOperations {
	return &FileOperations{
		db:                opts.DbSess,
		minMultiPartSize:  opts.MinMultiPartSize,
		externalFileMode:  opts.ExternalFileMode,
		externalFilesPath: opts.ExternalFilesPath,
		prefix:            opts.Prefix,
	}
}

func (f *FileOperations) CreateFile(req *CreateFileRequest, stream io.Reader) (int64, error) {
	exists, err := f.fileExists(req.OwnerID, req.Path, req.Name)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, fmt.Errorf("file already exists")
	}

	now := time.Now()
	fileMeta := &FileMeta{
		OwnerID:   req.OwnerID,
		Name:      req.Name,
		Path:      req.Path,
		StoreType: req.StoreType,
		CreatedBy: req.CreatedBy,
		IsFolder:  req.IsFolder,
		CreatedAt: &now,
		Size:      0,
	}

	if req.StoreType == 0 && !req.IsFolder {
		if f.externalFileMode {
			fileMeta.StoreType = StoreTypeExternal
		} else {
			fileMeta.StoreType = StoreTypeInline
		}
	}

	rid, err := f.fileMetaTable().Insert(fileMeta)
	if err != nil {
		return 0, err
	}
	fileID := rid.ID().(int64)

	defer func() {
		if err != nil {
			f.cleanupOnError(fileID)
		}
	}()

	if req.IsFolder {
		return fileID, nil
	}

	sizeTotal, hashSumStr, err := f.processFileContent(fileID, req, stream)
	if err != nil {
		return 0, err
	}

	err = f.fileMetaTable().Find(db.Cond{"id": fileID}).Update(map[string]any{
		"size": sizeTotal,
		"hash": hashSumStr,
	})

	return fileID, err
}

func (f *FileOperations) CreateFolder(ownerID int64, path string, name string, createdBy int64) (int64, error) {
	req := &CreateFileRequest{
		OwnerID:   ownerID,
		Name:      name,
		Path:      path,
		CreatedBy: createdBy,
		IsFolder:  true,
		StoreType: 0,
	}
	return f.CreateFile(req, nil)
}

func (f *FileOperations) GetFileMeta(id int64) (*FileMeta, error) {
	file := &FileMeta{}
	err := f.fileMetaTable().Find(db.Cond{"id": id}).One(file)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (f *FileOperations) GetFileMetaByPath(ownerID int64, path string, name string) (*FileMeta, error) {
	file := &FileMeta{}
	err := f.fileMetaTable().Find(db.Cond{
		"owner_id": ownerID,
		"path":     path,
		"name":     name,
	}).One(file)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (f *FileOperations) ListFiles(ownerID int64, path string) ([]FileMeta, error) {
	files := make([]FileMeta, 0)
	err := f.fileMetaTable().Find(db.Cond{
		"owner_id": ownerID,
		"path":     path,
	}).All(&files)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (f *FileOperations) GetFileContent(ownerID int64, id int64) ([]byte, error) {
	file, err := f.GetFileMeta(id)
	if err != nil {
		return nil, err
	}

	err = f.validateFileOwnership(file, ownerID)
	if err != nil {
		return nil, err
	}

	return f.getFileContentByMeta(file)
}

func (f *FileOperations) GetFileContentByPath(ownerID int64, path string, name string) ([]byte, error) {
	file, err := f.GetFileMetaByPath(ownerID, path, name)
	if err != nil {
		return nil, err
	}
	return f.getFileContentByMeta(file)
}

func (f *FileOperations) StreamFile(ownerID int64, id int64, w io.Writer) error {
	file, err := f.GetFileMeta(id)
	if err != nil {
		return err
	}

	err = f.validateFileOwnership(file, ownerID)
	if err != nil {
		return err
	}

	return f.streamFileByMeta(file, w)
}

func (f *FileOperations) StreamFileByPath(ownerID int64, path string, name string, w io.Writer) error {
	file, err := f.GetFileMetaByPath(ownerID, path, name)
	if err != nil {
		return err
	}
	return f.streamFileByMeta(file, w)
}

func (f *FileOperations) StreamFileToHTTP(ownerID int64, id int64, w http.ResponseWriter) error {
	file, err := f.GetFileMeta(id)
	if err != nil {
		return err
	}

	err = f.validateFileOwnership(file, ownerID)
	if err != nil {
		return err
	}

	f.setupHTTPHeaders(w, file)
	return f.streamFileByMeta(file, w)
}

func (f *FileOperations) UpdateFile(ownerID int64, id int64, stream io.Reader) error {
	file, err := f.GetFileMeta(id)
	if err != nil {
		return err
	}

	err = f.validateFileOwnership(file, ownerID)
	if err != nil {
		return err
	}

	return f.updateFileContent(file, stream)
}

func (f *FileOperations) RemoveFile(ownerID int64, id int64) error {
	file, err := f.GetFileMeta(id)
	if err != nil {
		return err
	}

	err = f.validateFileOwnership(file, ownerID)
	if err != nil {
		return err
	}

	return f.removeFileRecursively(ownerID, file)
}

func (f *FileOperations) UpdateFileMeta(ownerID int64, id int64, data map[string]any) error {
	return f.fileMetaTable().Find(db.Cond{
		"id":       id,
		"owner_id": ownerID,
	}).Update(data)
}
