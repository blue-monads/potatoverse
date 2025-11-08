package file

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

func (f *FileOperations) getFileContentByMeta(file *dbmodels.FileMeta) ([]byte, error) {
	switch file.StoreType {
	case StoreTypeInline:
		return f.getInlineBlob(file.ID)
	case StoreTypeExternal:
		buf := make([]byte, file.Size)
		bufWriter := bytes.NewBuffer(buf)
		err := f.getExternalFile(file, bufWriter)
		if err != nil {
			return nil, err
		}
		return buf, nil
	case StoreTypeMultipart:
		return f.getMultipartBlob(file.ID)
	default:
		return nil, fmt.Errorf("unknown storage type: %d", file.StoreType)
	}
}

func (f *FileOperations) streamFileByMeta(file *dbmodels.FileMeta, w io.Writer) error {
	switch file.StoreType {
	case StoreTypeInline:
		data, err := f.getInlineBlob(file.ID)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	case StoreTypeExternal:
		return f.getExternalFile(file, w)
	case StoreTypeMultipart:
		return f.streamMultipartBlob(file.ID, w)
	default:
		return fmt.Errorf("unknown storage type: %d", file.StoreType)
	}
}

func (f *FileOperations) processFileContent(ownerID, fileID int64, req *datahub.CreateFileRequest, stream io.Reader) (int64, string, error) {
	hash := sha1.New()
	sizeTotal := int64(0)

	switch f.storeType {
	case StoreTypeInline:
		size, err := f.storeInlineBlob(fileID, stream, hash)
		if err != nil {
			return 0, "", err
		}
		sizeTotal = size
	case StoreTypeExternal:
		size, err := f.storeExternalFile(fileID, ownerID, req.CreatedBy, req.Path, req.Name, stream, hash)
		if err != nil {
			return 0, "", err
		}
		sizeTotal = size
	case StoreTypeMultipart:
		size, err := f.storeMultipartBlob(fileID, stream, hash)
		if err != nil {
			return 0, "", err
		}
		sizeTotal = size
	default:
		return 0, "", fmt.Errorf("unknown storage type: %d", f.storeType)
	}

	hashSum := hash.Sum(nil)
	hashSumStr := base64.StdEncoding.EncodeToString(hashSum)

	return sizeTotal, hashSumStr, nil
}

func (f *FileOperations) updateFileContent(file *dbmodels.FileMeta, stream io.Reader) error {
	err := f.removeFileContent(file)
	if err != nil {
		return err
	}

	hash := sha1.New()
	sizeTotal := int64(0)

	switch file.StoreType {
	case StoreTypeInline:
		sizeTotal, err = f.storeInlineBlob(file.ID, stream, hash)
	case StoreTypeExternal:
		sizeTotal, err = f.storeExternalFile(file.ID, file.OwnerID, file.CreatedBy, file.Path, file.Name, stream, hash)
	case StoreTypeMultipart:
		sizeTotal, err = f.storeMultipartBlob(file.ID, stream, hash)
	default:
		return fmt.Errorf("unknown storage type: %d", file.StoreType)
	}

	if err != nil {
		return err
	}

	hashSum := hash.Sum(nil)
	hashSumStr := base64.StdEncoding.EncodeToString(hashSum)

	return f.fileMetaTable().Find(db.Cond{"id": file.ID}).Update(map[string]any{
		"size": sizeTotal,
		"hash": hashSumStr,
	})
}

func (f *FileOperations) removeFileRecursively(ownerID int64, file *dbmodels.FileMeta) error {
	if file.IsFolder {
		childFiles, err := f.ListFiles(ownerID, filepath.Join(file.Path, file.Name))
		if err == nil {
			for _, child := range childFiles {
				err = f.RemoveFile(ownerID, child.ID)
				if err != nil {
					return err
				}
			}
		}
	}

	err := f.removeFileContent(file)
	if err != nil {
		return err
	}

	return f.fileMetaTable().Find(db.Cond{"id": file.ID}).Delete()
}

func (f *FileOperations) validateFileOwnership(file *dbmodels.FileMeta, ownerID int64) error {
	if file.OwnerID != ownerID {
		return fmt.Errorf("file does not belong to the specified owner")
	}
	return nil
}

func (f *FileOperations) setupHTTPHeaders(w http.ResponseWriter, file *dbmodels.FileMeta) {
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))
}
