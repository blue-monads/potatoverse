package actions

import (
	"errors"
	"io"
	"net/http"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
)

func (c *Controller) ListSpaceFiles(spaceId int64, path string) ([]dbmodels.FileMeta, error) {
	fops := c.database.GetFileOps()
	files, err := fops.ListFiles(spaceId, path)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (c *Controller) GetSpaceFileByPath(spaceId int64, path, name string) (*dbmodels.FileMeta, error) {
	pFileOps := c.database.GetPackageFileOps()
	file, err := pFileOps.GetFileMetaByPath(spaceId, path, name)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (c *Controller) GetSpaceFile(spaceId int64, fileId int64) (*dbmodels.FileMeta, error) {

	fops := c.database.GetFileOps()
	file, err := fops.GetFileMeta(fileId)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (c *Controller) DownloadSpaceFile(spaceId, fileId int64, w http.ResponseWriter) error {
	err := c.validateSpaceFileOwnership(spaceId, fileId)
	if err != nil {
		return err
	}

	fops := c.database.GetFileOps()
	return fops.StreamFile(spaceId, fileId, w)

}

func (c *Controller) DeleteSpaceFile(spaceId, fileId int64) error {
	err := c.validateSpaceFileOwnership(spaceId, fileId)
	if err != nil {
		return err
	}

	fops := c.database.GetFileOps()
	return fops.RemoveFile(spaceId, fileId)

}

func (c *Controller) UploadSpaceFile(spaceId int64, name, path string, stream io.Reader, createdBy int64) (int64, error) {

	fops := c.database.GetFileOps()
	return fops.CreateFile(spaceId, &datahub.CreateFileRequest{
		Name:      name,
		Path:      path,
		CreatedBy: createdBy,
	}, stream)
}

func (c *Controller) CreateSpaceFolder(spaceId int64, name, path string, createdBy int64) (int64, error) {

	fops := c.database.GetFileOps()

	return fops.CreateFolder(spaceId, path, name, createdBy)
}

// private

func (c *Controller) validateSpaceFileOwnership(spaceId int64, fileId int64) error {
	file, err := c.GetSpaceFile(spaceId, fileId)
	if err != nil {
		return err
	}

	space, err := c.GetSpaceById(spaceId)
	if err != nil {
		return err
	}

	if file.OwnerID != space.InstalledId {
		return errors.New("file does not belong to the space")
	}

	return nil
}

func (c *Controller) GetSpaceById(id int64) (*dbmodels.Space, error) {
	return c.database.GetSpaceOps().GetSpace(id)
}
