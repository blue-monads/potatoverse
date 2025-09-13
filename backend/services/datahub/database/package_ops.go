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

func (d *DB) InstallPackage(file string) (int64, error) {
	zipFile, err := zip.OpenReader(file)
	if err != nil {
		return 0, err
	}
	defer zipFile.Close()

	packageJson := []byte{}
	for _, file := range zipFile.File {
		if file.Name == "package.json" {
			jsonFile, err := file.Open()
			if err != nil {
				return 0, err
			}
			defer jsonFile.Close()
			packageJson, _ = io.ReadAll(jsonFile)
			break
		}
	}

	pkg := &models.Package{}
	err = json.Unmarshal(packageJson, pkg)
	if err != nil {
		return 0, err
	}

	table := d.packagesTable()
	id, err := table.Insert(pkg)
	if err != nil {
		return 0, err
	}

	return id.ID().(int64), nil
}

func (d *DB) GetPackage(id int64) (*models.Package, error) {
	item := models.Package{}
	err := d.packagesTable().Find(db.Cond{"id": id}).One(&item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (d *DB) DeletePackage(id int64) error {
	return d.packagesTable().Find(db.Cond{"id": id}).Delete()
}

func (d *DB) UpdatePackage(id int64, data map[string]any) error {
	return d.packagesTable().Find(db.Cond{"id": id}).Update(data)
}

func (d *DB) ListPackages() ([]models.Package, error) {
	items := make([]models.Package, 0)
	err := d.packagesTable().Find().All(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (d *DB) ListPackageFiles(packageId int64) ([]models.PackageFile, error) {
	items := make([]models.PackageFile, 0)
	err := d.packageFilesTable().Find(db.Cond{"package_id": packageId}).All(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (d *DB) GetPackageFileMeta(packageId, id int64) (*models.PackageFile, error) {
	item := models.PackageFile{}
	err := d.packageFilesTable().Find(db.Cond{"package_id": packageId, "id": id}).One(&item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (d *DB) GetPackageFileMetaByPath(packageId int64, path, name string) (*models.PackageFile, error) {
	item := models.PackageFile{}
	err := d.packageFilesTable().Find(db.Cond{
		"package_id": packageId,
		"path":       path,
		"name":       name,
	}).One(&item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (d *DB) GetPackageFileStreaming(packageId, id int64, w io.Writer) error {
	item := models.PackageFile{}
	err := d.packageFilesTable().Find(db.Cond{"package_id": packageId, "id": id}).One(&item)
	if err != nil {
		return err
	}

	if item.StoreType != 1 {
		return fmt.Errorf("only external blobs are implemented for package files")
	}

	fileBlobs := make([]models.PackageFileBlobLite, 0)
	err = d.packageFileBlobsTable().Find(db.Cond{"file_id": item.ID}).All(&fileBlobs)
	if err != nil {
		return err
	}

	for _, fileBlob := range fileBlobs {
		blob, err := d.sess.SQL().Select("blob").From("PackageFileBlobs").Where(db.Cond{"id": fileBlob.ID}).QueryRow()
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

func (d *DB) GetPackageFile(packageId, id int64) ([]byte, error) {
	buf := bytes.Buffer{}
	err := d.GetPackageFileStreaming(packageId, id, &buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (d *DB) AddPackageFile(packageId int64, name string, path string, data []byte) (int64, error) {
	return d.AddPackageFileStreaming(packageId, name, path, bytes.NewReader(data))
}

func (d *DB) AddPackageFileStreaming(packageId int64, name string, path string, stream io.Reader) (int64, error) {
	t := time.Now()

	file := &models.PackageFile{
		PackageID: packageId,
		StoreType: 0,
		Name:      name,
		Path:      path,
		CreatedBy: 0,
		CreatedAt: &t,
	}

	rid, err := d.packageFilesTable().Insert(file)
	if err != nil {
		return 0, err
	}

	fileId := rid.ID().(int64)

	file.Size, err = d.setPackageBlobs(fileId, stream)
	if err != nil {
		return 0, err
	}

	err = d.packageFilesTable().Find(db.Cond{"id": fileId}).Update(map[string]any{"size": file.Size})
	if err != nil {
		return 0, err
	}

	return fileId, nil
}

func (d *DB) UpdatePackageFile(packageId, id int64, data []byte) error {
	return d.UpdatePackageFileStreaming(packageId, id, bytes.NewReader(data))
}

func (d *DB) UpdatePackageFileStreaming(packageId, id int64, stream io.Reader) error {
	err := d.cleanPackageBlobs(id)
	if err != nil {
		return err
	}

	size, err := d.setPackageBlobs(id, stream)
	if err != nil {
		return err
	}

	return d.packageFilesTable().Find(db.Cond{"id": id}).Update(map[string]any{"size": size})
}

func (d *DB) DeletePackageFile(packageId, id int64) error {
	err := d.cleanPackageBlobs(id)
	if err != nil {
		return err
	}
	return d.packageFilesTable().Find(db.Cond{"package_id": packageId, "id": id}).Delete()
}

func (d *DB) setPackageBlobs(fileId int64, stream io.Reader) (int64, error) {
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
		_, err = d.packageFileBlobsTable().Insert(models.PackageFileBlob{
			FileID: fileId,
			Size:   int64(n),
			PartID: int64(partId),
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

func (d *DB) cleanPackageBlobs(fileId int64) error {
	return d.packageFileBlobsTable().Find(db.Cond{"file_id": fileId}).Delete()
}

func (d *DB) packagesTable() db.Collection {
	return d.Table("Packages")
}

func (d *DB) packageFilesTable() db.Collection {
	return d.Table("PackageFiles")
}

func (d *DB) packageFileBlobsTable() db.Collection {
	return d.Table("PackageFileBlobs")
}
