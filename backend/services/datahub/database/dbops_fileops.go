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
	"github.com/blue-monads/turnix/backend/services/datahub/models"
	"github.com/jaevor/go-nanoid"
	"github.com/k0kubun/pp"

	"github.com/upper/db/v4"
)

var _ datahub.FileDataOps = (*DB)(nil)

func (d *DB) AddFolder(spaceId int64, uid int64, path string, name string) (int64, error) {
	t := time.Now()
	file := &models.File{
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

func (d *DB) AddFileStreaming(file *models.File, stream io.Reader) (id int64, err error) {
	pp.Println("@add_file_streaming/1", file.Path)
	filetable := d.filesTable()
	partstable := d.filePartedBlobsTable()
	exists, _ := filetable.Find(db.Cond{
		"name":           file.Name,
		"path":           file.Path,
		"created_by":     file.CreatedBy,
		"owner_space_id": file.OwnerSpaceID,
	}).Exists()
	pp.Println("@add_file_streaming/2", exists)
	if exists {
		pp.Println("@add_file_streaming/3")
		return 0, fmt.Errorf("file already exists")
	}
	if d.externalFileMode {
		file.StoreType = 1
	} else if file.Size > d.minFileMultiPartSize {
		file.StoreType = 2
	}
	pp.Println("@add_file_streaming/3", file.StoreType)
	rid, err := filetable.Insert(file)
	if err != nil {
		return 0, err
	}
	id = rid.ID().(int64)
	pp.Println("@add_file_streaming/4")
	defer func() {
		if err != nil {
			filetable.Find(db.Cond{"id": id}).Delete()
			partstable.Find(db.Cond{"file_id": id}).Delete()
		}
	}()
	pp.Println("@add_file_streaming/5")

	if d.externalFileMode {
		pp.Println("@add_file_streaming/7")
		f, err := os.Create(fmt.Sprintf("files/%d/%s", file.CreatedBy, file.Path))
		if err != nil {
			return 0, err
		}
		pp.Println("@add_file_streaming/8")
		defer f.Close()
		_, err = io.Copy(f, stream)
		if err != nil {
			return 0, err
		}
		pp.Println("@add_file_streaming/9")
		f.Sync()
		pp.Println("@add_file_streaming/10")
		return id, nil
	}
	driver := d.sess.Driver().(*sql.DB)
	if file.StoreType == 0 {
		pp.Println("@add_file_streaming/11")
		data, err := io.ReadAll(stream)
		if err != nil {
			return 0, err
		}
		pp.Println("@add_file_streaming/12")
		sizeTotal := int64(len(data))

		_, err = driver.Exec("UPDATE Files SET blob = ? , size = ? WHERE id = ?", data, sizeTotal, id)
		if err != nil {
			return 0, err
		}

		return id, nil
	} else if file.StoreType == 2 {
		pp.Println("@add_file_streaming/13")
		partId := 0
		buf := make([]byte, d.minFileMultiPartSize)
		sizeTotal := int64(0)
		for {
			pp.Println("@add_file_streaming/14", partId)
			n, err := stream.Read(buf)
			if err != nil && err != io.EOF {
				return 0, err
			}
			if n == 0 {
				break
			}

			sizeTotal += int64(n)
			pp.Println("@add_file_streaming/15", n)
			_, err = driver.Exec("INSERT INTO FilePartedBlobs (file_id, size, part_id, blob) VALUES (?, ?, ?, ?)", id, n, partId, buf[:n])
			if err != nil {
				return 0, err
			}
			pp.Println("@add_file_streaming/16")
			partId++
		}
		pp.Println("@add_file_streaming/17")

		err = d.filesTable().Find(db.Cond{"id": id}).Update(map[string]any{"size": sizeTotal})
		if err != nil {
			return 0, err
		}
	}
	pp.Println("@add_file_streaming/18")
	return id, nil
}

func (d *DB) GetFileBlobStreaming(id int64, w http.ResponseWriter) error {
	pp.Println("@get_file_blob_streaming/1", id)
	table := d.filesTable()
	file := models.File{}
	err := table.Find(db.Cond{"id": id}).One(&file)
	if err != nil {
		return err
	}
	pp.Println("@get_file_blob_streaming/2", file.CreatedBy, file.Path)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))
	switch file.StoreType {
	case 0:
		data := make([]byte, file.Size)
		pp.Println("@get_file_blob_streaming/3")
		row, err := d.sess.SQL().Select("blob").From("Files").Where(db.Cond{"id": id}).QueryRow()
		if err != nil {
			pp.Println("@get_file_blob_streaming/4", err)
			return err
		}
		pp.Println("@get_file_blob_streaming/5")
		err = row.Scan(&data)
		if err != nil {
			pp.Println("@get_file_blob_streaming/6", err)
			return err
		}
		pp.Println("@get_file_blob_streaming/7")
		written := int64(0)
		for written < file.Size || written == 0 {
			pp.Println("@get_file_blob_streaming/8", written, file.Size)
			n, err := w.Write(data[written:])
			if err != nil {
				return err
			}
			written += int64(n)
		}
		return err
	case 1:
		pp.Println("@get_file_blob_streaming/9")

		pp.Println("@get_file_blob_streaming/10", file.CreatedBy)
		out, err := os.ReadFile(fmt.Sprintf("files/%d/%s", file.CreatedBy, file.Path))
		if err != nil {
			return err
		}
		pp.Println("@get_file_blob_streaming/11")
		_, err = w.Write(out)
		if err != nil {
			return err
		}
	case 2:
		parts := make([]models.FilePartedBlob, 0)
		pp.Println("@get_file_blob_streaming/12")
		err := d.filePartedBlobsTable().Find(db.Cond{"file_id": id}).
			Select("id", "size", "part_id").
			OrderBy("part_id").
			All(&parts)
		if err != nil {
			pp.Println("@get_file_blob_streaming/13", err.Error())
			return err
		}
		pp.Println("@get_file_blob_streaming/14", len(parts))
		for _, part := range parts {
			pp.Println("@get_file_blob_streaming/15")
			blob, err := d.sess.SQL().Select("blob").From("FilePartedBlobs").Where(db.Cond{"id": part.Id}).QueryRow()
			if err != nil {
				pp.Println("@get_file_blob_streaming/16", err.Error())
				return err
			}
			pp.Println("@get_file_blob_streaming/17")
			data := make([]byte, part.Size)
			err = blob.Scan(&data)
			if err != nil {
				pp.Println("@get_file_blob_streaming/18", err.Error())
				return err
			}
			pp.Println("@get_file_blob_streaming/19")
			written := int64(0)
			for written <= part.Size {
				pp.Println("@get_file_blob_streaming/20")
				_, err = w.Write(data[written:])
				if err != nil {
					return err
				}
				written += int64(len(data))
			}
			pp.Println("@get_file_blob_streaming/21")
		}
		pp.Println("@get_file_blob_streaming/22")
		return nil
	}
	pp.Println("@get_file_blob_streaming/23")
	return nil
}
func (d *DB) GetFileMeta(id int64) (*models.File, error) {
	table := d.filesTable()
	file := models.File{}
	err := table.Find(db.Cond{"id": id}).One(&file)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (d *DB) ListFilesBySpace(spaceId int64, path string) ([]models.File, error) {
	table := d.filesTable()
	cond := db.Cond{
		"owner_space_id": spaceId,
		"path":           path,
	}

	pp.Println("@list_files_by_space/1", cond)

	files := make([]models.File, 0)
	err := table.Find(cond).All(&files)
	if err != nil {
		return nil, err
	}
	return files, nil
}
func (d *DB) ListFilesByUser(uid int64, path string) ([]models.File, error) {
	table := d.filesTable()
	cond := db.Cond{
		"ftype":         "user",
		"owner_user_id": uid,
		"path":          path,
	}
	files := make([]models.File, 0)
	err := table.Find(cond).All(&files)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (d *DB) RemoveFile(id int64) error {
	table := d.filesTable()
	file := models.File{}
	record := table.Find(db.Cond{"id": id})
	err := record.One(&file)
	if err != nil {
		return err
	}
	if file.IsFolder {
		// fixme => delete all files in folder
		pp.Println("@delete_files_form_folder/1")
	}
	switch file.StoreType {
	case 1:
		createdBy := file.CreatedBy
		err = os.Remove(fmt.Sprintf("files/%d/%s", createdBy, file.Path))
		if err != nil {
			return err
		}
	case 2:
		d.filePartedBlobsTable().Find(db.Cond{"file_id": id}).Delete()
	}
	return record.Delete()
}
func (d *DB) UpdateFile(id int64, data map[string]any) error {
	table := d.filesTable()
	return table.Find(db.Cond{"id": id}).Update(data)
}
func (d *DB) UpdateFileStreaming(file *models.File, stream io.Reader) (int64, error) {
	err := d.RemoveFile(file.ID)
	if err != nil {
		return 0, err
	}
	return d.AddFileStreaming(file, stream)
}

// share

func (d *DB) GetSharedFile(id string, w http.ResponseWriter) error {
	table := d.fileSharesTable()
	file := &models.FileShare{}
	err := table.Find(db.Cond{"id": id}).One(file)
	if err != nil {
		return err
	}
	return d.GetFileBlobStreaming(file.FileID, w)
}

var generator, _ = nanoid.CustomASCII("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 12)

func (d *DB) AddFileShare(fileId int64, userId int64, spaceId int64) (string, error) {
	file, err := d.GetFileMeta(fileId)
	if err != nil {
		return "", err
	}

	ext := filepath.Ext(file.Name)

	table := d.fileSharesTable()

	t := &time.Time{}

	shareId := fmt.Sprintf("%s%s", generator(), ext)

	data := &models.FileShare{
		ID:        shareId,
		FileID:    fileId,
		UserID:    userId,
		ProjectID: spaceId,
		CreatedAt: t,
	}

	pp.Println("@shareid", shareId)

	_, err = table.Insert(data)
	if err != nil {
		return "", err
	}
	return shareId, nil
}

func (d *DB) ListFileShares(fileId int64) ([]models.FileShare, error) {
	table := d.fileSharesTable()

	shares := make([]models.FileShare, 0)

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

func (d *DB) fileSharesTable() db.Collection {
	return d.Table("FileShares")
}

func (d *DB) filesTable() db.Collection {
	return d.Table("Files")
}

func (d *DB) filePartedBlobsTable() db.Collection {
	return d.Table("FilePartedBlobs")
}
