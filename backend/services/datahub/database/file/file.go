package file

import (
	"archive/zip"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/k0kubun/pp"

	"github.com/upper/db/v4"
)

//go:embed file_schema.sql
var FileSchemaSQL string

const BlobSizeLimit = 1024 * 1024 * 1

const (
	StoreTypeInline    = 0
	StoreTypeExternal  = 1
	StoreTypeMultipart = 2
)

type FileBlobLite struct {
	ID     int64 `db:"id"`
	FileID int64 `db:"file_id"`
	Size   int64 `db:"size"`
	PartID int64 `db:"part_id"`
}

type FileOperations struct {
	db                db.Session
	minMultiPartSize  int64
	externalFilesPath string
	prefix            string
	storeType         int64
}

type Options struct {
	DbSess            db.Session
	MinMultiPartSize  int64
	ExternalFilesPath string
	Prefix            string
	StoreType         int64
}

func NewFileOperations(opts Options) *FileOperations {
	return &FileOperations{
		db:                opts.DbSess,
		minMultiPartSize:  opts.MinMultiPartSize,
		externalFilesPath: opts.ExternalFilesPath,
		prefix:            opts.Prefix,
		storeType:         opts.StoreType,
	}
}

// "@fPath" "" "@fName" "potato.json"
// "@fPath" "public/" "@fName" "readme.txt"

func (f *FileOperations) ApplyZipToFile(ownerID int64, zipPath string) error {

	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}

	defer zipReader.Close()

	for _, file := range zipReader.File {
		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		fPath, fName := path.Split(file.Name)

		pp.Println("@fPath", fPath, "@fName", fName)

		fPath = strings.TrimPrefix(fPath, "/")
		fPath = strings.TrimSuffix(fPath, "/")

		if fPath == "" || fPath == "." {
			fPath = ""
		}

		req := &datahub.CreateFileRequest{
			Name:      fName,
			Path:      fPath,
			CreatedBy: ownerID,
		}
		_, err = f.CreateFile(ownerID, req, fileReader)
		if err != nil {
			return err
		}

		defer fileReader.Close()
	}

	return nil
}

func (f *FileOperations) CreateFile(ownerID int64, req *datahub.CreateFileRequest, stream io.Reader) (int64, error) {
	exists, err := f.fileExists(ownerID, req.Path, req.Name)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, fmt.Errorf("file already exists")
	}

	now := time.Now()

	fileMeta := &dbmodels.FileMeta{
		OwnerID:   ownerID,
		Name:      req.Name,
		Path:      req.Path,
		StoreType: f.storeType,
		CreatedBy: req.CreatedBy,
		IsFolder:  false,
		CreatedAt: &now,
		Size:      0,
		UpdatedBy: req.CreatedBy,
		UpdatedAt: &now,
	}

	fileMeta.StoreType = StoreTypeInline

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

	sizeTotal, hashSumStr, err := f.processFileContent(ownerID, fileID, req, stream)
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
	req := &datahub.CreateFileRequest{
		Name:      name,
		Path:      path,
		CreatedBy: createdBy,
	}
	return f.CreateFile(ownerID, req, nil)
}

func (f *FileOperations) GetFileMeta(id int64) (*dbmodels.FileMeta, error) {
	file := &dbmodels.FileMeta{}
	err := f.fileMetaTable().Find(db.Cond{"id": id}).One(file)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (f *FileOperations) GetFileMetaByPath(ownerID int64, path string, name string) (*dbmodels.FileMeta, error) {
	file := &dbmodels.FileMeta{}
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

func (f *FileOperations) ListFiles(ownerID int64, path string) ([]dbmodels.FileMeta, error) {
	files := make([]dbmodels.FileMeta, 0)
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

func (f *FileOperations) StreamFileToHTTP(ownerID int64, path, name string, w http.ResponseWriter) error {
	file, err := f.GetFileMetaByPath(ownerID, path, name)
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
