package actions

import (
	"errors"
	"io"
	"net/http"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/utils/qq"
)

func (c *Controller) ListSpaceFiles(installedId int64, path string) ([]dbmodels.FileMeta, error) {
	fops := c.database.GetFileOps()
	files, err := fops.ListFiles(installedId, path)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (c *Controller) GetSpaceFileByPath(installedId int64, path, name string) (*dbmodels.FileMeta, error) {
	pFileOps := c.database.GetPackageFileOps()
	file, err := pFileOps.GetFileMetaByPath(installedId, path, name)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (c *Controller) GetSpaceFile(installedId int64, fileId int64) (*dbmodels.FileMeta, error) {

	fops := c.database.GetFileOps()
	file, err := fops.GetFileMeta(fileId)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (c *Controller) DownloadSpaceFile(installedId int64, fileId int64, w http.ResponseWriter) error {
	qq.Println("@DownloadSpaceFile/1", installedId, fileId)
	err := c.validateSpaceFileOwnership(installedId, fileId)
	if err != nil {
		return err
	}

	qq.Println("@DownloadSpaceFile/2", "success")

	fops := c.database.GetFileOps()
	return fops.StreamFile(installedId, fileId, w)

}

func (c *Controller) DeleteSpaceFile(installedId int64, fileId int64) error {
	err := c.validateSpaceFileOwnership(installedId, fileId)
	if err != nil {
		return err
	}

	fops := c.database.GetFileOps()
	return fops.RemoveFile(installedId, fileId)

}

func (c *Controller) UploadSpaceFile(installedId int64, name, path string, stream io.Reader, createdBy int64) (int64, error) {

	fops := c.database.GetFileOps()
	return fops.CreateFile(installedId, &datahub.CreateFileRequest{
		Name:      name,
		Path:      path,
		CreatedBy: createdBy,
	}, stream)
}

func (c *Controller) CreateSpaceFolder(installedId int64, name, path string, createdBy int64) (int64, error) {

	fops := c.database.GetFileOps()

	return fops.CreateFolder(installedId, path, name, createdBy)
}

// ValidateSpaceFileOwnership validates that the file belongs to the specified space
func (c *Controller) validateSpaceFileOwnership(installedId int64, fileId int64) error {

	file, err := c.GetSpaceFile(installedId, fileId)
	if err != nil {
		return err
	}

	qq.Println("@validateSpaceFileOwnership/3", file.OwnerID)

	if file.OwnerID != installedId {
		return errors.New("file does not belong to the space")
	}

	qq.Println("@validateSpaceFileOwnership/4", "success")

	return nil
}

// private

func (c *Controller) GetSpaceById(id int64) (*dbmodels.Space, error) {
	return c.database.GetSpaceOps().GetSpace(id)
}
