package actions

import (
	"io"

	"github.com/blue-monads/turnix/backend/services/datahub/models"
)

func (c *Controller) ListPackageFiles(packageId int64) ([]models.PackageFile, error) {
	return c.database.ListPackageFiles(packageId)
}

func (c *Controller) GetPackageFile(packageId, fileId int64) (*models.PackageFile, error) {
	return c.database.GetPackageFileMeta(packageId, fileId)
}

func (c *Controller) DownloadPackageFile(packageId, fileId int64, w io.Writer) error {
	return c.database.GetPackageFileStreaming(packageId, fileId, w)
}

func (c *Controller) DeletePackageFile(packageId, fileId int64) error {
	return c.database.DeletePackageFile(packageId, fileId)
}

func (c *Controller) UploadPackageFile(packageId int64, name, path string, stream io.Reader) (int64, error) {
	return c.database.AddPackageFileStreaming(packageId, name, path, stream)
}
