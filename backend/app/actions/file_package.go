package actions

import (
	"io"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
)

func (c *Controller) ListPackageFiles(packageId int64, path string) ([]dbmodels.FileMeta, error) {

	files, err := c.database.GetPackageFileOps().ListFiles(packageId, path)

	if err != nil {
		return nil, err
	}

	return files, nil
}

func (c *Controller) GetPackageFile(packageId, fileId int64) (*dbmodels.FileMeta, error) {
	return c.database.GetPackageFileOps().GetFileMeta(fileId)
}

func (c *Controller) DownloadPackageFile(packageId, fileId int64, w io.Writer) error {
	return c.database.GetPackageFileOps().StreamFile(packageId, fileId, w)

}

func (c *Controller) DeletePackageFile(packageId, fileId int64) error {
	return c.database.GetPackageFileOps().RemoveFile(packageId, fileId)
}

func (c *Controller) UploadPackageFile(packageId, createdBy int64, name, path string, stream io.Reader) (int64, error) {
	fileId, err := c.database.GetPackageFileOps().CreateFile(packageId,
		&datahub.CreateFileRequest{
			Name:      name,
			Path:      path,
			CreatedBy: createdBy,
		}, stream)

	if err != nil {
		return 0, err
	}

	return fileId, nil
}
