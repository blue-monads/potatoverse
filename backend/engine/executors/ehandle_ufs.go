package executors

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	- share_file

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
		// Strip leading slash to make it relative
		after = strings.TrimPrefix(after, "/")
		return "home", after
	} else if after, ok := strings.CutPrefix(path, "/pkg"); ok {
		// Strip leading slash to make it relative
		after = strings.TrimPrefix(after, "/")
		return "pkg", after
	} else if after0, ok0 := strings.CutPrefix(path, "/tmp"); ok0 {
		// Strip leading slash to make it relative
		after0 = strings.TrimPrefix(after0, "/")
		return "tmp", after0
	}

	return "unknown", path
}

func isRootPath(path string) bool {
	return path == "/"
}

func shouldStartWithSlash(path string) bool {
	return strings.HasPrefix(path, "/")
}

func (c *EHandle) ListFiles(path string) ([]File, error) {
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
		// Normalize root path
		if cleanPath == "/" || cleanPath == "." {
			cleanPath = ""
		}

		files, err := c.Database.ListSpaceFiles(c.SpaceId, cleanPath)
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
		// Normalize root path: database stores root files with empty string path
		if cleanPath == "/" || cleanPath == "." {
			cleanPath = ""
		}

		files, err := c.Database.ListPackageFilesByPath(c.PackageId, cleanPath)
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
		// Use "." for root of tmp directory
		if cleanPath == "" {
			cleanPath = "."
		}

		dir, err := c.FsRoot.Open(cleanPath)
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
				UpdatedAt: entry.ModTime(),
			}
		}
		return files, nil
	default:
		return nil, errors.New("unknown backend")
	}
}

func (c *EHandle) ReadFile(path string) ([]byte, error) {
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

		// Normalize root path: filepath.Dir returns "/" or "." for root files
		// but database stores root files with empty string path
		if filePath == "/" || filePath == "." {
			filePath = ""
		}

		// Get file metadata by path and name
		file, err := c.Database.GetSpaceFileMetaByPathAndName(c.SpaceId, filePath, fileName)
		if err != nil {
			return nil, err
		}

		if file.IsFolder {
			return nil, errors.New("path is a directory")
		}

		return c.Database.GetSpaceFile(c.SpaceId, file.ID)

	case "pkg":
		fileName := filepath.Base(cleanPath)
		filePath := filepath.Dir(cleanPath)

		// Normalize root path: filepath.Dir returns "/" or "." for root files
		// but database stores root files with empty string path
		if filePath == "/" || filePath == "." {
			filePath = ""
		}

		// Check if it's a directory first
		file, err := c.Database.GetPackageFileMetaByPath(c.PackageId, filePath, fileName)
		if err != nil {
			return nil, err
		}
		if file.IsFolder {
			return nil, errors.New("path is a directory")
		}

		// Read package file content directly using path
		var buf bytes.Buffer
		err = c.Database.GetPackageFileStreamingByPath(c.PackageId, filePath, fileName, &buf)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil

	case "tmp":
		file, err := c.FsRoot.Open(cleanPath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return io.ReadAll(file)

	default:
		return nil, errors.New("unknown backend")
	}
}

func (c *EHandle) WriteFile(uid int64, path string, data []byte) error {
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

		// Normalize root path
		if filePath == "/" || filePath == "." {
			filePath = ""
		}

		_, err := c.Database.StreamAddSpaceFile(c.SpaceId, 0, filePath, fileName, bytes.NewReader(data))
		return err

	case "pkg":
		return errors.New("package file writing not implemented - packages are read-only")

	case "tmp":
		file, err := c.FsRoot.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
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

func (c *EHandle) RemoveFile(uid int64, path string) error {

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

		// Normalize root path
		if filePath == "/" || filePath == "." {
			filePath = ""
		}

		// Find the file by path and name
		file, err := c.Database.GetSpaceFileMetaByPathAndName(c.SpaceId, filePath, fileName)
		if err != nil {
			return err
		}

		return c.Database.RemoveSpaceFile(c.SpaceId, file.ID)

	case "pkg":
		return errors.New("package file removal not implemented - packages are read-only")

	case "tmp":
		return c.FsRoot.Remove(cleanPath)

	default:
		return errors.New("unknown backend")
	}
}

func (c *EHandle) Mkdir(uid int64, path string) error {
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

		// Normalize root path
		if filePath == "/" || filePath == "." {
			filePath = ""
		}

		// Create folder in database
		_, err := c.Database.AddSpaceFolder(c.SpaceId, 0, filePath, fileName)
		return err

	case "pkg":
		return errors.New("package directory creation not implemented - packages are read-only")

	case "tmp":
		return c.FsRoot.Mkdir(cleanPath, 0755)
	default:
		return errors.New("unknown backend")
	}
}

func (c *EHandle) Rmdir(uid int64, path string) error {
	if !shouldStartWithSlash(path) {
		return errors.New("path must start with /")
	}

	backend, cleanPath := backend(path)

	switch backend {
	case "home":
		fileName := filepath.Base(cleanPath)
		filePath := filepath.Dir(cleanPath)

		// Normalize root path
		if filePath == "/" || filePath == "." {
			filePath = ""
		}

		file, err := c.Database.GetSpaceFileMetaByPathAndName(c.SpaceId, filePath, fileName)
		if err != nil {
			return err
		}

		if !file.IsFolder {
			return errors.New("path is not a directory")
		}

		return c.Database.RemoveSpaceFile(c.SpaceId, file.ID)

	case "pkg":
		return errors.New("package directory removal not implemented - packages are read-only")

	case "tmp":
		return c.FsRoot.Remove(cleanPath)

	default:
		return errors.New("unknown backend")
	}
}

func (c *EHandle) Exists(path string) (bool, error) {
	if !shouldStartWithSlash(path) {
		return false, errors.New("path must start with /")
	}

	backend, cleanPath := backend(path)

	switch backend {
	case "home":
		fileName := filepath.Base(cleanPath)
		filePath := filepath.Dir(cleanPath)

		// Normalize root path
		if filePath == "/" || filePath == "." {
			filePath = ""
		}

		// Check if file exists in database
		_, err := c.Database.GetSpaceFileMetaByPathAndName(c.SpaceId, filePath, fileName)
		if err != nil {
			return false, nil
		}
		return true, nil

	case "pkg":
		fileName := filepath.Base(cleanPath)
		filePath := filepath.Dir(cleanPath)

		// Normalize root path: filepath.Dir returns "/" or "." for root files
		// but database stores root files with empty string path
		if filePath == "/" || filePath == "." {
			filePath = ""
		}

		_, err := c.Database.GetPackageFileMetaByPath(c.PackageId, filePath, fileName)
		if err != nil {
			return false, nil // File doesn't exist
		}
		return true, nil

	case "tmp":
		// Use "." for root of tmp directory
		if cleanPath == "" {
			cleanPath = "."
		}
		_, err := c.FsRoot.Stat(cleanPath)
		return !os.IsNotExist(err), nil

	default:
		return false, errors.New("unknown backend")
	}
}

func (c *EHandle) ShareFile(uid int64, path string) (string, error) {
	if !shouldStartWithSlash(path) {
		return "", errors.New("path must start with /")
	}

	backend, cleanPath := backend(path)
	if backend != "home" {
		return "", errors.New("You can only share files in the home")
	}

	fileName := filepath.Base(cleanPath)
	filePath := filepath.Dir(cleanPath)

	// Normalize root path
	if filePath == "/" || filePath == "." {
		filePath = ""
	}

	file, err := c.Database.GetSpaceFileMetaByPathAndName(c.SpaceId, filePath, fileName)
	if err != nil {
		return "", err
	}

	return c.Database.AddFileShare(file.ID, uid, c.SpaceId)
}
