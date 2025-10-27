package actions

import (
	"io"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
)

func (c *Controller) ListPackageFiles(packageId int64, path string) ([]dbmodels.FileMeta, error) {
	return []dbmodels.FileMeta{}, nil
}

func (c *Controller) GetPackageFile(packageId, fileId int64) (*dbmodels.FileMeta, error) {
	return nil, nil
}

func (c *Controller) DownloadPackageFile(packageId, fileId int64, w io.Writer) error {
	return nil
}

func (c *Controller) DeletePackageFile(packageId, fileId int64) error {
	return nil
}

func (c *Controller) UploadPackageFile(packageId int64, name, path string, stream io.Reader) (int64, error) {
	return 0, nil
}
