package actions

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blue-monads/turnix/backend/services/datahub"
)

/*

/home -> SpaceFileOps
/pkg -> PackageOps
/tmp -> LocalFsModule(root *os.Root)



	- list
	- read_file
	- write_file
	- remove_file
	- mkdir
	- rmdir
	- exists

*/

type File struct {
	Name      string    `json:"name"`
	IsFolder  bool      `json:"is_folder"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func backend(path string) (string, string) {
	if after, ok := strings.CutPrefix(path, "/home"); ok {
		return "home", after
	} else if after, ok := strings.CutPrefix(path, "/pkg"); ok {
		return "pkg", after
	} else if after0, ok0 := strings.CutPrefix(path, "/tmp"); ok0 {
		return "tmp", after0
	}

	return "tmp", path
}

type UnifiedFs struct {
	database  datahub.Database
	root      *os.Root
	packageId int64
}

func isRootPath(path string) bool {
	return path == "/"
}

func shouldStartWithSlash(path string) bool {
	return strings.HasPrefix(path, "/")
}

func (c *UnifiedFs) ListFiles(spaceId int64, path string) ([]File, error) {
	if !shouldStartWithSlash(path) {
		return nil, errors.New("path must start with /")
	}

	if isRootPath(path) {
		return []File{
			{
				Name:     "home",
				IsFolder: true,
				Size:     0,
			},
			{
				Name:     "pkg",
				IsFolder: true,
				Size:     0,
			},
			{
				Name:     "tmp",
				IsFolder: true,
				Size:     0,
			},
		}, nil
	}

	backend, cleanPath := backend(path)

	switch backend {
	case "home":
		files, err := c.database.ListSpaceFiles(spaceId, cleanPath)
		if err != nil {
			return nil, err
		}
		rFiles := make([]File, len(files))
		for i, file := range files {
			rFiles[i] = File{
				Name:      file.Name,
				IsFolder:  file.IsFolder,
				Size:      file.Size,
				CreatedAt: *file.CreatedAt,
				UpdatedAt: *file.CreatedAt,
			}
		}
		return rFiles, nil
	case "pkg":
		files, err := c.database.ListPackageFilesByPath(c.packageId, cleanPath)
		if err != nil {
			return nil, err
		}
		rFiles := make([]File, len(files))
		for i, file := range files {
			rFiles[i] = File{
				Name:      file.Name,
				IsFolder:  file.IsFolder,
				Size:      file.Size,
				CreatedAt: *file.CreatedAt,
			}
		}
		return rFiles, nil
	case "tmp":
		dir, err := c.root.Open(cleanPath)
		if err != nil {
			return nil, err
		}
		defer dir.Close()

		entries, err := dir.Readdir(-1)
		if err != nil {
			return nil, err
		}

		files := make([]File, len(entries))
		for i, entry := range entries {
			files[i] = File{
				Name:      entry.Name(),
				IsFolder:  entry.IsDir(),
				Size:      entry.Size(),
				CreatedAt: time.Now(), // os.FileInfo doesn't provide creation time
				UpdatedAt: entry.ModTime(),
			}
		}
		return files, nil
	default:
		return nil, errors.New("unknown backend")
	}
}

func (c *UnifiedFs) ReadFile(spaceId int64, path string) ([]byte, error) {
	if !shouldStartWithSlash(path) {
		return nil, errors.New("path must start with /")
	}

	if isRootPath(path) {
		return nil, errors.New("cannot read root path")
	}

	backend, cleanPath := backend(path)

	switch backend {
	case "home":
		fileName := filepath.Base(cleanPath)
		filePath := filepath.Dir(cleanPath)

		// Get file metadata by path
		file, err := c.database.GetSpaceFileMetaByPath(spaceId, filepath.Join(filePath, fileName))
		if err != nil {
			return nil, err
		}

		if file.IsFolder {
			return nil, errors.New("path is a directory")
		}

		return c.database.GetSpaceFile(spaceId, file.ID)

	case "pkg":
		fileName := filepath.Base(cleanPath)
		filePath := filepath.Dir(cleanPath)

		// Check if it's a directory first
		file, err := c.database.GetPackageFileMetaByPath(c.packageId, filePath, fileName)
		if err != nil {
			return nil, err
		}
		if file.IsFolder {
			return nil, errors.New("path is a directory")
		}

		// Read package file content directly using path
		var buf bytes.Buffer
		err = c.database.GetPackageFileStreamingByPath(c.packageId, filePath, fileName, &buf)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil

	case "tmp":
		file, err := c.root.Open(cleanPath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return io.ReadAll(file)

	default:
		return nil, errors.New("unknown backend")
	}
}

func (c *UnifiedFs) WriteFile(spaceId int64, path string, data []byte) error {
	if !shouldStartWithSlash(path) {
		return errors.New("path must start with /")
	}

	if isRootPath(path) {
		return errors.New("cannot write to root path")
	}

	backend, cleanPath := backend(path)

	switch backend {
	case "home":
		fileName := filepath.Base(cleanPath)
		filePath := filepath.Dir(cleanPath)

		_, err := c.database.StreamAddSpaceFile(spaceId, 0, filePath, fileName, bytes.NewReader(data))
		return err

	case "pkg":
		return errors.New("package file writing not implemented - packages are read-only")

	case "tmp":
		file, err := c.root.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = file.Write(data)
		return err

	default:
		return errors.New("unknown backend")
	}
}

func (c *UnifiedFs) RemoveFile(spaceId int64, path string) error {

	if !shouldStartWithSlash(path) {
		return errors.New("path must start with /")
	}

	if isRootPath(path) {
		return errors.New("cannot remove root path")
	}

	backend, cleanPath := backend(path)

	switch backend {
	case "home":
		fileName := filepath.Base(cleanPath)
		filePath := filepath.Dir(cleanPath)

		// Find the file by path and name
		file, err := c.database.GetSpaceFileMetaByPath(spaceId, filepath.Join(filePath, fileName))
		if err != nil {
			// Try alternative: find by listing files
			files, err := c.database.ListSpaceFiles(spaceId, filePath)
			if err != nil {
				return err
			}
			// Find the file by name
			found := false
			for _, f := range files {
				if f.Name == fileName {
					file = &f
					found = true
					break
				}
			}
			if !found {
				return errors.New("file not found")
			}
		}

		return c.database.RemoveSpaceFile(spaceId, file.ID)

	case "pkg":
		return errors.New("package file removal not implemented - packages are read-only")

	case "tmp":
		return c.root.Remove(cleanPath)

	default:
		return errors.New("unknown backend")
	}
}

func (c *UnifiedFs) Mkdir(spaceId int64, path string) error {
	if !shouldStartWithSlash(path) {
		return errors.New("path must start with /")
	}

	if isRootPath(path) {
		return errors.New("cannot create directory in root path")
	}

	backend, cleanPath := backend(path)

	switch backend {
	case "home":
		fileName := filepath.Base(cleanPath)
		filePath := filepath.Dir(cleanPath)

		// Create folder in database
		_, err := c.database.AddSpaceFolder(spaceId, 0, filePath, fileName)
		return err

	case "pkg":
		return errors.New("package directory creation not implemented - packages are read-only")

	case "tmp":
		// Use local filesystem
		if c.root == nil {
			return errors.New("local filesystem not available")
		}
		return c.root.Mkdir(cleanPath, 0755)

	default:
		return errors.New("unknown backend")
	}
}

func (c *UnifiedFs) Rmdir(spaceId int64, path string) error {
	if !shouldStartWithSlash(path) {
		return errors.New("path must start with /")
	}

	backend, cleanPath := backend(path)

	switch backend {
	case "home":
		fileName := filepath.Base(cleanPath)
		filePath := filepath.Dir(cleanPath)

		file, err := c.database.GetSpaceFileMetaByPath(spaceId, filepath.Join(filePath, fileName))
		if err != nil {
			// Try alternative: find by listing
			files, err := c.database.ListSpaceFiles(spaceId, filePath)
			if err != nil {
				return err
			}
			found := false
			for _, f := range files {
				if f.Name == fileName {
					file = &f
					found = true
					break
				}
			}
			if !found {
				return errors.New("directory not found")
			}
		}

		if !file.IsFolder {
			return errors.New("path is not a directory")
		}

		return c.database.RemoveSpaceFile(spaceId, file.ID)

	case "pkg":
		return errors.New("package directory removal not implemented - packages are read-only")

	case "tmp":
		return c.root.Remove(cleanPath)

	default:
		return errors.New("unknown backend")
	}
}

func (c *UnifiedFs) Exists(spaceId int64, path string) (bool, error) {
	if !shouldStartWithSlash(path) {
		return false, errors.New("path must start with /")
	}

	backend, cleanPath := backend(path)

	switch backend {
	case "home":
		fileName := filepath.Base(cleanPath)
		filePath := filepath.Dir(cleanPath)

		// Check if file exists in database
		_, err := c.database.GetSpaceFileMetaByPath(spaceId, filepath.Join(filePath, fileName))
		if err != nil {
			// Try alternative: find by listing
			files, err := c.database.ListSpaceFiles(spaceId, filePath)
			if err != nil {
				return false, nil
			}
			// Check if file exists in list
			for _, f := range files {
				if f.Name == fileName {
					return true, nil
				}
			}
			return false, nil
		}
		return true, nil

	case "pkg":

		fileName := filepath.Base(cleanPath)
		filePath := filepath.Dir(cleanPath)
		_, err := c.database.GetPackageFileMetaByPath(c.packageId, filePath, fileName)
		if err != nil {
			return false, nil // File doesn't exist
		}
		return true, nil

	case "tmp":
		_, err := c.root.Stat(cleanPath)
		return !os.IsNotExist(err), nil

	default:
		return false, errors.New("unknown backend")
	}
}
