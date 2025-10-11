package database

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/jaevor/go-nanoid"
	"github.com/k0kubun/pp"

	"github.com/upper/db/v4"
)

var _ datahub.SpaceFileOps = (*DB)(nil)

// StreamAddSpaceFile adds a file to a space with streaming support
func (d *DB) StreamAddSpaceFile(spaceId int64, uid int64, path string, name string, stream io.Reader) (id int64, err error) {
	pp.Println("@stream_add_space_file/1", spaceId, uid, path, name)

	t := time.Now()
	file := &dbmodels.File{
		Name:         name,
		Path:         path,
		CreatedBy:    uid,
		OwnerSpaceID: spaceId,
		Size:         0,
		CreatedAt:    &t,
		IsFolder:     false,
	}

	filetable := d.filesTable()
	partstable := d.filePartedBlobsTable()

	// Check if file already exists
	exists, _ := filetable.Find(db.Cond{
		"name":           file.Name,
		"path":           file.Path,
		"created_by":     file.CreatedBy,
		"owner_space_id": file.OwnerSpaceID,
	}).Exists()

	pp.Println("@stream_add_space_file/2", exists)
	if exists {
		return 0, fmt.Errorf("file already exists")
	}

	if d.externalFileMode {
		file.StoreType = 1
	} else if file.Size > d.minFileMultiPartSize {
		file.StoreType = 2
	}

	pp.Println("@stream_add_space_file/3", file.StoreType)
	rid, err := filetable.Insert(file)
	if err != nil {
		return 0, err
	}
	id = rid.ID().(int64)

	pp.Println("@stream_add_space_file/4", id)

	defer func() {
		if err != nil {
			filetable.Find(db.Cond{"id": id}).Delete()
			partstable.Find(db.Cond{"file_id": id}).Delete()
		}
	}()

	if d.externalFileMode {
		pp.Println("@stream_add_space_file/5")
		dirPath := fmt.Sprintf("files/%d/%d", spaceId, uid)
		os.MkdirAll(dirPath, 0755)

		filePath := filepath.Join(dirPath, path, name)
		os.MkdirAll(filepath.Dir(filePath), 0755)

		f, err := os.Create(filePath)
		if err != nil {
			return 0, err
		}
		defer f.Close()

		written, err := io.Copy(f, stream)
		if err != nil {
			return 0, err
		}

		f.Sync()

		// Update size
		err = filetable.Find(db.Cond{"id": id}).Update(map[string]any{"size": written})
		if err != nil {
			return 0, err
		}

		return id, nil
	}

	driver := d.sess.Driver().(*sql.DB)

	if file.StoreType == 0 {
		pp.Println("@stream_add_space_file/6")
		data, err := io.ReadAll(stream)
		if err != nil {
			return 0, err
		}

		sizeTotal := int64(len(data))
		_, err = driver.Exec("UPDATE Files SET blob = ? , size = ? WHERE id = ?", data, sizeTotal, id)
		if err != nil {
			return 0, err
		}

		return id, nil
	} else if file.StoreType == 2 {
		pp.Println("@stream_add_space_file/7")
		partId := 0
		buf := make([]byte, d.minFileMultiPartSize)
		sizeTotal := int64(0)

		for {
			n, err := stream.Read(buf)
			if err != nil && err != io.EOF {
				return 0, err
			}
			if n == 0 {
				break
			}

			sizeTotal += int64(n)
			_, err = driver.Exec("INSERT INTO FilePartedBlobs (file_id, size, part_id, blob) VALUES (?, ?, ?, ?)", id, n, partId, buf[:n])
			if err != nil {
				return 0, err
			}
			partId++
		}

		err = filetable.Find(db.Cond{"id": id}).Update(map[string]any{"size": sizeTotal})
		if err != nil {
			return 0, err
		}
	}

	pp.Println("@stream_add_space_file/8")
	return id, nil
}

// AddSpaceFolder creates a folder in a space
func (d *DB) AddSpaceFolder(spaceId int64, uid int64, path string, name string) (int64, error) {
	t := time.Now()
	file := &dbmodels.File{
		Name:         name,
		Path:         path,
		CreatedBy:    uid,
		OwnerSpaceID: spaceId,
		Size:         0,
		CreatedAt:    &t,
		IsFolder:     true,
	}

	table := d.filesTable()
	rid, err := table.Insert(file)
	if err != nil {
		return 0, err
	}
	id := rid.ID().(int64)
	return id, nil
}

// GetSpaceFileMetaByPath retrieves file metadata by path
func (d *DB) GetSpaceFileMetaByPath(spaceId int64, path string) (*dbmodels.File, error) {
	table := d.filesTable()
	file := &dbmodels.File{}
	err := table.Find(db.Cond{
		"owner_space_id": spaceId,
		"path":           path,
	}).One(file)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// GetSpaceFileMetaById retrieves file metadata by ID
func (d *DB) GetSpaceFileMetaById(id int64) (*dbmodels.File, error) {
	table := d.filesTable()
	file := &dbmodels.File{}
	err := table.Find(db.Cond{"id": id}).One(file)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// GetSpaceFile retrieves file content as bytes
func (d *DB) GetSpaceFile(spaceId int64, id int64) ([]byte, error) {
	file, err := d.GetSpaceFileMetaById(id)
	if err != nil {
		return nil, err
	}

	if file.OwnerSpaceID != spaceId {
		return nil, fmt.Errorf("file does not belong to the specified space")
	}

	switch file.StoreType {
	case 0:
		data := make([]byte, file.Size)
		row, err := d.sess.SQL().Select("blob").From("Files").Where(db.Cond{"id": id}).QueryRow()
		if err != nil {
			return nil, err
		}
		err = row.Scan(&data)
		if err != nil {
			return nil, err
		}
		return data, nil
	case 1:
		filePath := fmt.Sprintf("files/%d/%d/%s/%s", spaceId, file.CreatedBy, file.Path, file.Name)
		return os.ReadFile(filePath)
	case 2:
		parts := make([]dbmodels.FilePartedBlob, 0)
		err := d.filePartedBlobsTable().Find(db.Cond{"file_id": id}).
			Select("id", "size", "part_id").
			OrderBy("part_id").
			All(&parts)
		if err != nil {
			return nil, err
		}

		result := make([]byte, 0, file.Size)
		for _, part := range parts {
			blob, err := d.sess.SQL().Select("blob").From("FilePartedBlobs").Where(db.Cond{"id": part.Id}).QueryRow()
			if err != nil {
				return nil, err
			}
			data := make([]byte, part.Size)
			err = blob.Scan(&data)
			if err != nil {
				return nil, err
			}
			result = append(result, data...)
		}
		return result, nil
	}

	return nil, fmt.Errorf("unknown store type")
}

// StreamGetSpaceFile streams file content to http.ResponseWriter
func (d *DB) StreamGetSpaceFile(spaceId int64, uid int64, id int64, w http.ResponseWriter) error {
	pp.Println("@stream_get_space_file/1", spaceId, uid, id)

	file, err := d.GetSpaceFileMetaById(id)
	if err != nil {
		return err
	}

	if file.OwnerSpaceID != spaceId {
		return fmt.Errorf("file does not belong to the specified space")
	}

	pp.Println("@stream_get_space_file/2", file.StoreType)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))

	switch file.StoreType {
	case 0:
		data := make([]byte, file.Size)
		row, err := d.sess.SQL().Select("blob").From("Files").Where(db.Cond{"id": id}).QueryRow()
		if err != nil {
			return err
		}
		err = row.Scan(&data)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	case 1:
		filePath := fmt.Sprintf("files/%d/%d/%s/%s", spaceId, file.CreatedBy, file.Path, file.Name)
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	case 2:
		parts := make([]dbmodels.FilePartedBlob, 0)
		err := d.filePartedBlobsTable().Find(db.Cond{"file_id": id}).
			Select("id", "size", "part_id").
			OrderBy("part_id").
			All(&parts)
		if err != nil {
			return err
		}

		for _, part := range parts {
			blob, err := d.sess.SQL().Select("blob").From("FilePartedBlobs").Where(db.Cond{"id": part.Id}).QueryRow()
			if err != nil {
				return err
			}
			data := make([]byte, part.Size)
			err = blob.Scan(&data)
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

	return fmt.Errorf("unknown store type")
}

// StreamGetSpaceFileByPath streams file content by path
func (d *DB) StreamGetSpaceFileByPath(spaceId int64, uid int64, path string, name string, w http.ResponseWriter) error {
	pp.Println("@stream_get_space_file_by_path/1", spaceId, uid, path, name)

	table := d.filesTable()
	file := &dbmodels.File{}

	fullPath := filepath.Join(path, name)

	err := table.Find(db.Cond{
		"owner_space_id": spaceId,
		"path":           fullPath,
	}).One(file)

	if err != nil {
		// Try alternative: path and name separately
		err = table.Find(db.Cond{
			"owner_space_id": spaceId,
			"path":           path,
			"name":           name,
		}).One(file)
		if err != nil {
			return err
		}
	}

	return d.StreamGetSpaceFile(spaceId, uid, file.ID, w)
}

// RemoveSpaceFile deletes a file from a space
func (d *DB) RemoveSpaceFile(spaceId, id int64) error {
	table := d.filesTable()
	file := &dbmodels.File{}
	record := table.Find(db.Cond{"id": id, "owner_space_id": spaceId})

	err := record.One(file)
	if err != nil {
		return err
	}

	if file.IsFolder {
		// Delete all files in folder
		pp.Println("@remove_space_file/delete_folder", file.Path)
		folderPath := filepath.Join(file.Path, file.Name)
		childFiles := make([]dbmodels.File, 0)
		err = table.Find(db.Cond{
			"owner_space_id": spaceId,
			"path":           folderPath,
		}).All(&childFiles)
		if err == nil {
			for _, child := range childFiles {
				d.RemoveSpaceFile(spaceId, child.ID)
			}
		}
	}

	switch file.StoreType {
	case 1:
		filePath := fmt.Sprintf("files/%d/%d/%s/%s", spaceId, file.CreatedBy, file.Path, file.Name)
		err = os.Remove(filePath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	case 2:
		d.filePartedBlobsTable().Find(db.Cond{"file_id": id}).Delete()
	}

	return record.Delete()
}

// UpdateSpaceFile updates file metadata
func (d *DB) UpdateSpaceFile(spaceId, id int64, data map[string]any) error {
	table := d.filesTable()
	return table.Find(db.Cond{"id": id, "owner_space_id": spaceId}).Update(data)
}

// StreamUpdateSpaceFile updates file content with streaming
func (d *DB) StreamUpdateSpaceFile(spaceId, id int64, stream io.Reader) (int64, error) {
	// Get the file metadata first
	file, err := d.GetSpaceFileMetaById(id)
	if err != nil {
		return 0, err
	}

	if file.OwnerSpaceID != spaceId {
		return 0, fmt.Errorf("file does not belong to the specified space")
	}

	// Remove old file
	err = d.RemoveSpaceFile(spaceId, id)
	if err != nil {
		return 0, err
	}

	// Add new file with same metadata
	return d.StreamAddSpaceFile(spaceId, file.CreatedBy, file.Path, file.Name, stream)
}

func (d *DB) ListSpaceFiles(spaceId int64, path string) ([]dbmodels.File, error) {
	table := d.filesTable()
	cond := db.Cond{
		"owner_space_id": spaceId,
		"path":           path,
	}

	pp.Println("@list_space_files/1", cond)

	files := make([]dbmodels.File, 0)
	err := table.Find(cond).All(&files)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// File Share Operations

func (d *DB) GetSharedFile(id string, w http.ResponseWriter) error {
	table := d.fileSharesTable()
	share := &dbmodels.FileShare{}
	err := table.Find(db.Cond{"id": id}).One(share)
	if err != nil {
		return err
	}

	file, err := d.GetSpaceFileMetaById(share.FileID)
	if err != nil {
		return err
	}

	return d.StreamGetSpaceFile(file.OwnerSpaceID, share.UserID, share.FileID, w)
}

var generator, _ = nanoid.CustomASCII("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 12)

func (d *DB) AddFileShare(fileId int64, userId int64, spaceId int64) (string, error) {
	file, err := d.GetSpaceFileMetaById(fileId)
	if err != nil {
		return "", err
	}

	if file.OwnerSpaceID != spaceId {
		return "", fmt.Errorf("file does not belong to the specified space")
	}

	ext := filepath.Ext(file.Name)
	table := d.fileSharesTable()
	t := time.Now()
	shareId := fmt.Sprintf("%s%s", generator(), ext)

	data := &dbmodels.FileShare{
		ID:        shareId,
		FileID:    fileId,
		UserID:    userId,
		ProjectID: spaceId,
		CreatedAt: &t,
	}

	pp.Println("@add_file_share/shareid", shareId)

	_, err = table.Insert(data)
	if err != nil {
		return "", err
	}
	return shareId, nil
}

func (d *DB) ListFileShares(fileId int64) ([]dbmodels.FileShare, error) {
	table := d.fileSharesTable()
	shares := make([]dbmodels.FileShare, 0)

	err := table.Find(db.Cond{"file_id": fileId}).All(&shares)
	if err != nil {
		return nil, err
	}

	return shares, nil
}

func (d *DB) RemoveFileShare(userId int64, id string) error {
	table := d.fileSharesTable()
	return table.Find(db.Cond{"id": id, "user_id": userId}).Delete()
}

// Helper methods

func (d *DB) fileSharesTable() db.Collection {
	return d.Table("FileShares")
}

func (d *DB) filesTable() db.Collection {
	return d.Table("Files")
}

func (d *DB) filePartedBlobsTable() db.Collection {
	return d.Table("FilePartedBlobs")
}
