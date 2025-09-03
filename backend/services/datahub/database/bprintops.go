package database

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/blue-monads/turnix/backend/services/datahub/models"
	"github.com/upper/db/v4"
)

func (d *DB) CreateBprintInstall(file string) (int64, error) {

	zipFile, err := zip.OpenReader(file)
	if err != nil {
		return 0, err
	}
	defer zipFile.Close()

	bprintJson := []byte{}

	for _, file := range zipFile.File {
		if file.Name == "bprint.json" {
			jsonFile, err := file.Open()
			if err != nil {
				return 0, err
			}
			defer jsonFile.Close()
			jsonFile.Read(bprintJson)
			break
		}
	}

	bprintInstall := &models.BprintInstall{}
	err = json.Unmarshal(bprintJson, bprintInstall)
	if err != nil {
		return 0, err
	}

	table := d.bprintInstallsTable()
	id, err := table.Insert(bprintInstall)
	if err != nil {
		return 0, err
	}

	return id.ID().(int64), nil
}

func (d *DB) AddBprintFileStreaming(slug, name, path string, stream io.Reader) (int64, error) {
	t := time.Now()

	file := &models.BprintInstallFile{
		BprintSlug: slug,
		StoreType:  1,
		Name:       name,
		Path:       path,
		CreatedBy:  0,
		CreatedAt:  &t,
	}

	rid, err := d.bprintInstallFilesTable().Insert(file)
	if err != nil {
		return 0, err
	}

	fileId := rid.ID().(int64)

	file.Size, err = d.setBprintBlobs(fileId, stream)
	if err != nil {
		return 0, err
	}

	err = d.bprintInstallFilesTable().Find(db.Cond{"id": fileId}).Update(db.Cond{"size": file.Size})
	if err != nil {
		return 0, err
	}

	return fileId, nil
}

func (d *DB) AddBprintFile(slug, name, path string, data []byte) (int64, error) {
	return d.AddBprintFileStreaming(slug, name, path, bytes.NewReader(data))
}

func (d *DB) ListBprintInstalls() ([]models.BprintInstall, error) {
	items := make([]models.BprintInstall, 0)
	err := d.bprintInstallsTable().Find().All(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (d *DB) GetBprintInstall(id int64) (*models.BprintInstall, error) {
	item := models.BprintInstall{}
	err := d.bprintInstallsTable().Find(db.Cond{"id": id}).One(&item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (d *DB) DeleteBprintInstall(id int64) error {
	return d.bprintInstallsTable().Find(db.Cond{"id": id}).Delete()
}

type Item struct {
	Id int64 `db:"id"`
}

func (d *DB) UpdateBprintFile(slug, name, path string, data []byte) error {
	return d.UpdateBprintFileStreaming(slug, path, bytes.NewReader(data))
}

func (d *DB) UpdateBprintFileStreaming(slug, path string, stream io.Reader) error {
	item := Item{}
	err := d.bprintInstallFilesTable().Find(db.Cond{"bprint_slug": slug, "path": path}).One(&item)
	if err != nil {
		return err
	}

	fileId := item.Id

	err = d.cleanBprintBlobs(fileId)
	if err != nil {
		return err
	}

	fileId, err = d.setBprintBlobs(fileId, stream)
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) ListBprintRootFiles(slug string) ([]models.BprintInstallFile, error) {
	items := make([]models.BprintInstallFile, 0)
	err := d.bprintInstallFilesTable().Find(db.Cond{"bprint_slug": slug, "is_folder": false}).All(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (d *DB) ListBprintFolderFiles(slug string, path string) ([]models.BprintInstallFile, error) {
	items := make([]models.BprintInstallFile, 0)
	err := d.bprintInstallFilesTable().Find(db.Cond{"bprint_slug": slug, "is_folder": true, "path": path}).All(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (d *DB) GetBprintFileMeta(slug string, path string) (*models.BprintInstallFile, error) {
	item := models.BprintInstallFile{}
	err := d.bprintInstallFilesTable().Find(db.Cond{"bprint_slug": slug, "path": path}).One(&item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (d *DB) GetBprintFileBlobStreaming(slug string, path string, w io.Writer) error {
	item := models.BprintInstallFile{}
	err := d.bprintInstallFilesTable().Find(db.Cond{"bprint_slug": slug, "path": path}).One(&item)
	if err != nil {
		return err
	}

	if item.StoreType != 1 {
		return fmt.Errorf("only external blobs are implemented for bprint files")
	}

	fileBlobs := make([]models.BprintInstallFileBlobLite, 0)
	err = d.bprintInstallFileBlobsTable().Find(db.Cond{"file_id": item.ID}).All(&fileBlobs)
	if err != nil {
		return err
	}

	for _, fileBlob := range fileBlobs {
		blob, err := d.sess.SQL().Select("blob").From("BprintInstallFileBlobs").Where(db.Cond{"id": fileBlob.ID}).QueryRow()
		if err != nil {
			return err
		}
		blobBytes := make([]byte, fileBlob.Size)
		err = blob.Scan(&blobBytes)
		if err != nil {
			return err
		}
		_, err = w.Write(blobBytes)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DB) GetBprintFile(slug string, path string) ([]byte, error) {
	item := models.BprintInstallFile{}
	err := d.bprintInstallFilesTable().Find(db.Cond{"bprint_slug": slug, "path": path}).One(&item)
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}

	err = d.GetBprintFileBlobStreaming(slug, path, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (d *DB) RemoveBprintFile(slug string, path string) error {
	item := Item{}
	err := d.bprintInstallFilesTable().Find(db.Cond{"bprint_slug": slug, "path": path}).One(&item)
	if err != nil {
		return err
	}
	err = d.cleanBprintBlobs(item.Id)
	if err != nil {
		return err
	}

	return d.bprintInstallFilesTable().Find(db.Cond{"id": item.Id}).Delete()
}

// private

func (d *DB) setBprintBlobs(fileId int64, stream io.Reader) (int64, error) {

	buf := make([]byte, d.minFileMultiPartSize)
	partId := 0
	totalSize := int64(0)
	for {
		n, err := stream.Read(buf)
		if err != nil && err != io.EOF {
			return 0, err
		}
		if n == 0 {
			break
		}
		_, err = d.bprintInstallFileBlobsTable().Insert(models.BprintInstallFileBlob{
			FileId: fileId,
			Size:   int64(n),
			PartId: int64(partId),
			Blob:   buf[:n],
		})
		if err != nil {
			return 0, err
		}
		partId++
		totalSize += int64(n)
	}

	return totalSize, nil

}

func (d *DB) cleanBprintBlobs(fileId int64) error {
	return d.bprintInstallFileBlobsTable().Find(db.Cond{"file_id": fileId}).Delete()
}

func (d *DB) bprintInstallsTable() db.Collection {
	return d.Table("BprintInstalls")
}

func (d *DB) bprintInstallFilesTable() db.Collection {
	return d.Table("BprintInstallFiles")
}

func (d *DB) bprintInstallFileBlobsTable() db.Collection {
	return d.Table("BprintInstallFileBlobs")
}
