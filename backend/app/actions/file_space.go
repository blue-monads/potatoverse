package actions

import (
	"io"
	"net/http"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
)

func (c *Controller) ListSpaceFiles(spaceId int64, path string) ([]dbmodels.FileMeta, error) {
	return []dbmodels.FileMeta{}, nil
}

func (c *Controller) GetSpaceFileByPath(spaceId int64, path, name string) (*dbmodels.FileMeta, error) {
	pFileOps := c.database.GetPackageFileOps()
	file, err := pFileOps.GetFileMetaByPath(spaceId, path, name)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (c *Controller) GetSpaceFile(spaceId, fileId int64) (*dbmodels.FileMeta, error) {
	// First get the file to verify it belongs to the space
	return nil, nil
}

func (c *Controller) DownloadSpaceFile(spaceId, fileId int64, w http.ResponseWriter) error {
	return nil
}

func (c *Controller) DeleteSpaceFile(spaceId, fileId int64) error {
	// First verify the file belongs to the space
	_, err := c.GetSpaceFile(spaceId, fileId)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) UploadSpaceFile(spaceId int64, name, path string, stream io.Reader, createdBy int64) (int64, error) {
	return 0, nil
}

func (c *Controller) CreateSpaceFolder(spaceId int64, name, path string, createdBy int64) (int64, error) {
	return 0, nil
}
