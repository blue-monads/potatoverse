package file

import (
	"database/sql"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"

	"github.com/upper/db/v4"
)

func (f *FileOperations) storeInlineBlob(fileID int64, stream io.Reader, hash hash.Hash) (int64, error) {
	data, err := io.ReadAll(stream)
	if err != nil {
		return 0, err
	}

	hash.Write(data)
	sizeTotal := int64(len(data))

	driver := f.db.Driver().(*sql.DB)
	_, err = driver.Exec("UPDATE "+f.getTableName()+" SET blob = ?, size = ? WHERE id = ?", data, sizeTotal, fileID)
	if err != nil {
		return 0, err
	}

	return sizeTotal, nil
}

func (f *FileOperations) storeExternalFile(fileID int64, ownerID int64, createdBy int64, path string, name string, stream io.Reader, hash hash.Hash) (int64, error) {
	dirPath := fmt.Sprintf("%s/%d/%d", f.externalFilesPath, ownerID, createdBy)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return 0, err
	}

	filePath := filepath.Join(dirPath, path, name)
	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return 0, err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	written, err := io.Copy(file, stream)
	if err != nil {
		return 0, err
	}

	file.Sync()

	err = f.readFileHash(file, hash)
	if err != nil {
		return 0, err
	}

	return written, nil
}

func (f *FileOperations) storeMultipartBlob(fileID int64, stream io.Reader, hash hash.Hash) (int64, error) {
	partID := 0
	sizeTotal := int64(0)
	buf := make([]byte, f.minMultiPartSize)

	for {
		n, err := stream.Read(buf)
		if err != nil && err != io.EOF {
			return 0, err
		}
		if n == 0 {
			break
		}

		currBytes := buf[:n]
		hash.Write(currBytes)

		_, err = f.fileBlobTable().Insert(FileBlob{
			FileID: fileID,
			Size:   int64(n),
			PartID: int64(partID),
			Blob:   currBytes,
		})
		if err != nil {
			return 0, err
		}

		sizeTotal += int64(n)
		partID++
	}

	return sizeTotal, nil
}

func (f *FileOperations) getInlineBlob(fileID int64) ([]byte, error) {
	row, err := f.db.SQL().Select("blob").From(f.getTableName()).Where(db.Cond{"id": fileID}).QueryRow()
	if err != nil {
		return nil, err
	}

	var data []byte
	err = row.Scan(&data)
	return data, err
}

func (f *FileOperations) getMultipartBlob(fileID int64) ([]byte, error) {
	parts := make([]FileBlobLite, 0)
	err := f.fileBlobTable().Find(db.Cond{"file_id": fileID}).
		Select("id", "size", "part_id").
		OrderBy("part_id").
		All(&parts)
	if err != nil {
		return nil, err
	}

	result := make([]byte, 0)
	for _, part := range parts {
		row, err := f.db.SQL().Select("blob").From(f.getBlobTableName()).Where(db.Cond{"id": part.ID}).QueryRow()
		if err != nil {
			return nil, err
		}

		data := make([]byte, part.Size)
		err = row.Scan(&data)
		if err != nil {
			return nil, err
		}

		result = append(result, data...)
	}

	return result, nil
}

func (f *FileOperations) streamMultipartBlob(fileID int64, w io.Writer) error {
	parts := make([]FileBlobLite, 0)
	err := f.fileBlobTable().Find(db.Cond{"file_id": fileID}).
		Select("id", "size", "part_id").
		OrderBy("part_id").
		All(&parts)
	if err != nil {
		return err
	}

	for _, part := range parts {
		row, err := f.db.SQL().Select("blob").From(f.getBlobTableName()).Where(db.Cond{"id": part.ID}).QueryRow()
		if err != nil {
			return err
		}

		data := make([]byte, part.Size)
		err = row.Scan(&data)
		if err != nil {
			return err
		}

		_, err = w.Write(data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FileOperations) getExternalFile(file *FileMeta, w io.Writer) error {
	filePath := fmt.Sprintf("%s/%d/%d/%s/%s", f.externalFilesPath, file.OwnerID, file.CreatedBy, file.Path, file.Name)
	ofile, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer ofile.Close()

	_, err = io.Copy(w, ofile)
	if err != nil {
		return err
	}

	return nil
}

func (f *FileOperations) removeFileContent(file *FileMeta) error {
	switch file.StoreType {
	case StoreTypeExternal:
		filePath := fmt.Sprintf("%s/%d/%d/%s/%s", f.externalFilesPath, file.OwnerID, file.CreatedBy, file.Path, file.Name)
		err := os.Remove(filePath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	case StoreTypeMultipart:
		return f.fileBlobTable().Find(db.Cond{"file_id": file.ID}).Delete()
	}
	return nil
}
