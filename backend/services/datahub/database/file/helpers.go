package file

import (
	"hash"
	"io"
	"os"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/upper/db/v4"
)

func (f *FileOperations) fileExists(ownerID int64, path string, name string) (bool, error) {

	cond := db.Cond{
		"owner_id": ownerID,
		"path":     path,
		"name":     name,
	}

	qq.Println("@fileExists/1", cond)

	return f.fileMetaTable().Find(cond).Exists()
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
		if n > 0 {
			hash.Write(buf[:n])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	return nil
}

func (f *FileOperations) cleanupOnError(fileID int64) {
	f.fileMetaTable().Find(db.Cond{"id": fileID}).Delete()
	f.fileBlobTable().Find(db.Cond{"file_id": fileID}).Delete()
}
