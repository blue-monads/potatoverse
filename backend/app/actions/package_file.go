package actions

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/blue-monads/turnix/backend/services/datahub/models"
)

func (c *Controller) ListPackageFiles(packageId int64, path string) ([]models.PackageFile, error) {
	return c.database.ListPackageFilesByPath(packageId, path)
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

// Space KV operations

func (c *Controller) GetSpace(spaceId int64) (*models.Space, error) {
	return c.database.GetSpace(spaceId)
}

func (c *Controller) QuerySpaceKV(spaceId int64, cond map[any]any) ([]models.SpaceKV, error) {
	return c.database.QuerySpaceKV(spaceId, cond)
}

func (c *Controller) GetSpaceKVByID(spaceId, kvId int64) (*models.SpaceKV, error) {
	// First get all KV entries for the space and find by ID
	kvEntries, err := c.database.QuerySpaceKV(spaceId, map[any]any{})
	if err != nil {
		return nil, err
	}

	for _, kv := range kvEntries {
		if kv.ID == kvId {
			return &kv, nil
		}
	}

	return nil, errors.New("KV entry not found")
}

func (c *Controller) CreateSpaceKV(spaceId int64, data map[string]any) (*models.SpaceKV, error) {
	// Validate required fields
	key, ok := data["key"].(string)
	if !ok || key == "" {
		return nil, errors.New("key is required")
	}

	groupName, ok := data["group_name"].(string)
	if !ok || groupName == "" {
		return nil, errors.New("group_name is required")
	}

	value, ok := data["value"].(string)
	if !ok || value == "" {
		return nil, errors.New("value is required")
	}

	// Extract optional fields
	tag1, _ := data["tag1"].(string)
	tag2, _ := data["tag2"].(string)
	tag3, _ := data["tag3"].(string)

	kv := &models.SpaceKV{
		SpaceID: spaceId,
		Key:     key,
		Group:   groupName,
		Value:   value,
		Tag1:    tag1,
		Tag2:    tag2,
		Tag3:    tag3,
	}

	err := c.database.AddSpaceKV(spaceId, kv)
	if err != nil {
		return nil, err
	}

	// Get the created KV entry
	return c.database.GetSpaceKV(spaceId, groupName, key)
}

func (c *Controller) UpdateSpaceKVByID(spaceId, kvId int64, data map[string]any) (*models.SpaceKV, error) {
	// Get the existing KV entry
	kv, err := c.GetSpaceKVByID(spaceId, kvId)
	if err != nil {
		return nil, err
	}

	// Update using group and key
	err = c.database.UpdateSpaceKV(spaceId, kv.Group, kv.Key, data)
	if err != nil {
		return nil, err
	}

	// Return updated entry
	return c.database.GetSpaceKV(spaceId, kv.Group, kv.Key)
}

func (c *Controller) DeleteSpaceKVByID(spaceId, kvId int64) error {
	// Get the existing KV entry to find group and key
	kv, err := c.GetSpaceKVByID(spaceId, kvId)
	if err != nil {
		return err
	}

	return c.database.RemoveSpaceKV(spaceId, kv.Group, kv.Key)
}

// Space Files operations

func (c *Controller) ListSpaceFiles(spaceId int64, path string) ([]models.File, error) {
	return c.database.ListFilesBySpace(spaceId, path)
}

func (c *Controller) GetSpaceFile(spaceId, fileId int64) (*models.File, error) {
	// First get the file to verify it belongs to the space
	file, err := c.database.GetFileMeta(fileId)
	if err != nil {
		return nil, err
	}

	if file.OwnerSpaceID != spaceId {
		return nil, errors.New("file does not belong to this space")
	}

	return file, nil
}

func (c *Controller) DownloadSpaceFile(spaceId, fileId int64, w http.ResponseWriter) error {
	_, err := c.GetSpaceFile(spaceId, fileId)
	if err != nil {
		return err
	}

	// Use the existing file streaming method
	return c.database.GetFileBlobStreaming(fileId, w)
}

func (c *Controller) DeleteSpaceFile(spaceId, fileId int64) error {
	// First verify the file belongs to the space
	_, err := c.GetSpaceFile(spaceId, fileId)
	if err != nil {
		return err
	}

	return c.database.RemoveFile(fileId)
}

func (c *Controller) UploadSpaceFile(spaceId int64, name, path string, stream io.Reader, createdBy int64) (int64, error) {
	// Create file metadata
	now := time.Now()
	file := &models.File{
		Name:         name,
		Path:         path,
		OwnerSpaceID: spaceId,
		CreatedBy:    createdBy,
		IsFolder:     false,
		Size:         0, // Will be set by AddFileStreaming
		CreatedAt:    &now,
	}

	return c.database.AddFileStreaming(file, stream)
}

func (c *Controller) CreateSpaceFolder(spaceId int64, name, path string, createdBy int64) (int64, error) {
	return c.database.AddFolder(spaceId, createdBy, path, name)
}
