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
	return f.db.Collection(f.getTableName())
}

func (f *FileOperations) fileBlobTable() db.Collection {
	return f.db.Collection(f.getBlobTableName())
}

func (f *FileOperations) getTableName() string {
	return f.prefix + "FileMeta"
}

func (f *FileOperations) getBlobTableName() string {
	return f.prefix + "FileBlob"
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
