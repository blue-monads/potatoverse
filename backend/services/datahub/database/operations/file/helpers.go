package file

import (
	"hash"
	"os"

	"github.com/upper/db/v4"
)

func (f *FileOperations) fileExists(ownerID int64, path string, name string) (bool, error) {
	return f.fileMetaTable().Find(db.Cond{
		"owner_id": ownerID,
		"path":     path,
		"name":     name,
	}).Exists()
}

func (f *FileOperations) fileMetaTable() db.Collection {
	if f.context.Type == "space" {
		return f.db.Collection("Files")
	}
	return f.db.Collection("PackageFiles")
}

func (f *FileOperations) fileBlobTable() db.Collection {
	if f.context.Type == "space" {
		return f.db.Collection("FilePartedBlobs")
	}
	return f.db.Collection("PackageFileBlobs")
}

func (f *FileOperations) getTableName() string {
	if f.context.Type == "space" {
		return "Files"
	}
	return "PackageFiles"
}

func (f *FileOperations) getBlobTableName() string {
	if f.context.Type == "space" {
		return "FilePartedBlobs"
	}
	return "PackageFileBlobs"
}

func (f *FileOperations) readFileHash(file *os.File, hash hash.Hash) error {
	buf := make([]byte, 1024*1024)
	for {
		n, err := file.Read(buf)
		if err != nil {
			return err
		}

		if n == 0 {
			break
		}
		hash.Write(buf[:n])
	}

	return nil
}

func (f *FileOperations) cleanupOnError(fileID int64) {
	f.fileMetaTable().Find(db.Cond{"id": fileID}).Delete()
	f.fileBlobTable().Find(db.Cond{"file_id": fileID}).Delete()
}
