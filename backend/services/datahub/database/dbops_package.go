package database

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/k0kubun/pp"
	"github.com/upper/db/v4"
)

// "@file" "public/"
// "@file" "public/index.html"
// "@file" "turnix.json"

func (d *DB) InstallPackage(userId int64, file string) (int64, error) {
	zipFile, err := zip.OpenReader(file)
	if err != nil {
		return 0, err
	}
	defer zipFile.Close()

	packageJson := []byte{}
	for _, file := range zipFile.File {
		pp.Println("@file", file.Name)
		if file.Name == "turnix.json" || file.Name == "/turnix.json" {
			jsonFile, err := file.Open()
			if err != nil {
				return 0, err
			}
			defer jsonFile.Close()
			packageJson, _ = io.ReadAll(jsonFile)
			break
		}
	}

	pkg := &dbmodels.Package{}
	err = json.Unmarshal(packageJson, pkg)
	if err != nil {
		pp.Println("@packageJson", string(packageJson))
		pp.Println("@err/1", err)
		return 0, err
	}

	pkg.InstalledBy = userId

	table := d.packagesTable()
	id, err := table.Insert(pkg)
	if err != nil {
		pp.Println("@err/2", err)
		return 0, err
	}

	packageId := id.ID().(int64)

	folderToCreate := make(map[string]bool)

	for _, file := range zipFile.File {
		nameParts := strings.Split(file.Name, "/")
		fpath := strings.Join(nameParts[:len(nameParts)-1], "/")
		fname := nameParts[len(nameParts)-1]

		if strings.HasSuffix(file.Name, "/") {

			fid, err := d.createPackageFolder(packageId, fname, fpath)
			if err != nil {
				return 0, err
			}

			pp.Println("@fid", fid)

			continue
		}

		fileData, err := file.Open()
		if err != nil {
			return 0, err
		}
		defer fileData.Close()

		folderToCreate[fpath] = true

		fid, err := d.AddPackageFileStreaming(packageId, fname, fpath, fileData)
		if err != nil {
			return 0, err
		}

		pp.Println("@fid", fid)

	}

	for folderPath := range folderToCreate {
		err := d.createParentFoldersRecursively(packageId, folderPath)
		if err != nil {
			return 0, err
		}
	}

	return id.ID().(int64), nil
}

func (d *DB) GetPackage(id int64) (*dbmodels.Package, error) {
	item := dbmodels.Package{}
	err := d.packagesTable().Find(db.Cond{"id": id}).One(&item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (d *DB) DeletePackage(id int64) error {

	files, err := d.ListPackageFiles(id)
	if err != nil {
		return err
	}
	for _, file := range files {
		err = d.DeletePackageFile(id, file.ID)
		if err != nil {
			return err
		}
	}

	return d.packagesTable().Find(db.Cond{"id": id}).Delete()
}

func (d *DB) UpdatePackage(id int64, data map[string]any) error {
	return d.packagesTable().Find(db.Cond{"id": id}).Update(data)
}

func (d *DB) ListPackages() ([]dbmodels.Package, error) {
	items := make([]dbmodels.Package, 0)
	err := d.packagesTable().Find().All(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (d *DB) ListPackagesByIds(ids []int64) ([]dbmodels.Package, error) {
	items := make([]dbmodels.Package, 0)
	err := d.packagesTable().Find(db.Cond{"id": ids}).All(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (d *DB) ListPackageFilesByPath(packageId int64, path string) ([]dbmodels.PackageFile, error) {
	items := make([]dbmodels.PackageFile, 0)
	err := d.packageFilesTable().Find(db.Cond{"package_id": packageId, "path": path}).All(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (d *DB) ListPackageFiles(packageId int64) ([]dbmodels.PackageFile, error) {
	items := make([]dbmodels.PackageFile, 0)

	cond := db.Cond{"package_id": packageId}

	err := d.packageFilesTable().Find(cond).All(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (d *DB) GetPackageFileMeta(packageId, id int64) (*dbmodels.PackageFile, error) {
	item := dbmodels.PackageFile{}
	err := d.packageFilesTable().Find(db.Cond{"package_id": packageId, "id": id}).One(&item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (d *DB) GetPackageFileMetaByPath(packageId int64, path, name string) (*dbmodels.PackageFile, error) {
	item := dbmodels.PackageFile{}
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

type Ref struct {
	ID int64 `json:"id" db:"id,omitempty"`
}

func (d *DB) GetPackageFileStreamingByPath(packageId int64, path, name string, w io.Writer) error {
	ref := Ref{}
	err := d.packageFilesTable().Find(db.Cond{
		"package_id": packageId,
		"path":       path,
		"name":       name,
	}).One(&ref)
	if err != nil {
		return err
	}

	return d.GetPackageFileStreaming(packageId, ref.ID, w)
}

func (d *DB) GetPackageFileStreaming(packageId, id int64, w io.Writer) error {
	pp.Println("@GetPackageFileStreaming/1")

	item := dbmodels.PackageFile{}
	err := d.packageFilesTable().Find(db.Cond{"package_id": packageId, "id": id}).One(&item)
	if err != nil {
		pp.Println("@GetPackageFileStreaming/2", err)
		return err
	}
	pp.Println("@GetPackageFileStreaming/3")

	if item.StoreType != 2 {
		pp.Println("@GetPackageFileStreaming/4")
		return fmt.Errorf("only external blobs are implemented for package files")
	}
	pp.Println("@GetPackageFileStreaming/5")

	fileBlobs := make([]dbmodels.PackageFileBlobLite, 0)
	err = d.packageFileBlobsTable().Find(db.Cond{"file_id": item.ID}).All(&fileBlobs)
	if err != nil {
		pp.Println("@GetPackageFileStreaming/6", err)
		return err
	}
	pp.Println("@GetPackageFileStreaming/7")

	pp.Println("@fileBlobs", len(fileBlobs))

	for _, fileBlob := range fileBlobs {
		pp.Println("@GetPackageFileStreaming/8", fileBlob.ID)
		blob, err := d.sess.SQL().Select("blob").From("PackageFileBlobs").Where(db.Cond{"id": fileBlob.ID}).OrderBy("part_id").QueryRow()
		if err != nil {
			pp.Println("@GetPackageFileStreaming/9", err)
			return err
		}

		pp.Println("@GetPackageFileStreaming/10")

		blobBytes := make([]byte, fileBlob.Size)

		pp.Println("@GetPackageFileStreaming/11")

		err = blob.Scan(&blobBytes)
		pp.Println("@GetPackageFileStreaming/12", err)

		if err != nil {
			pp.Println("@GetPackageFileStreaming/13", err)
			return err
		}
		pp.Println("@GetPackageFileStreaming/14")
		_, err = w.Write(blobBytes)
		if err != nil {
			pp.Println("@GetPackageFileStreaming/15", err)
			return err
		}

		pp.Println("@GetPackageFileStreaming/16")
	}

	pp.Println("@GetPackageFileStreaming/17/end")

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

func (d *DB) createPackageFolder(packageId int64, path, name string) (int64, error) {
	t := time.Now()
	file := &dbmodels.PackageFile{
		PackageID: packageId,
		StoreType: 0,
		Name:      name,
		IsFolder:  true,
		Path:      path,
		CreatedBy: 0,
		CreatedAt: &t,
		Size:      0,
	}

	rid, err := d.packageFilesTable().Insert(file)
	if err != nil {
		return 0, err
	}
	return rid.ID().(int64), nil
}

func (d *DB) createParentFoldersRecursively(packageId int64, folderPath string) error {
	if folderPath == "" {
		return nil
	}

	pathParts := strings.Split(folderPath, "/")

	currentPath := ""
	for i, part := range pathParts {
		if part == "" {
			continue
		}

		// Build the current path
		if currentPath == "" {
			currentPath = part
		} else {
			currentPath = currentPath + "/" + part
		}

		// Create the parent path for this folder
		var parentPath string
		if i == 0 {
			parentPath = ""
		} else {
			parentParts := pathParts[:i]
			parentPath = strings.Join(parentParts, "/")
		}

		// Check if this folder already exists
		_, err := d.GetPackageFileMetaByPath(packageId, parentPath, part)
		if err == nil {
			// Folder already exists, continue to next level
			continue
		}

		// Create the folder
		_, err = d.createPackageFolder(packageId, parentPath, part)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DB) AddPackageFile(packageId int64, name string, path string, data []byte) (int64, error) {
	return d.AddPackageFileStreaming(packageId, name, path, bytes.NewReader(data))
}

func (d *DB) AddPackageFileStreaming(packageId int64, name string, path string, stream io.Reader) (int64, error) {
	t := time.Now()

	file := &dbmodels.PackageFile{
		PackageID: packageId,
		StoreType: 2,
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
		_, err = d.packageFileBlobsTable().Insert(dbmodels.PackageFileBlob{
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
