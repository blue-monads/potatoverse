package xutils

import (
	"archive/zip"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

func NewTempFile(path string, file *os.File) *TempFile {
	file.Seek(0, 0)
	return &TempFile{
		path:       path,
		name:       file.Name(),
		ReadCloser: file,
	}
}

type TempFile struct {
	io.ReadCloser
	path string
	name string
}

func (zr *TempFile) Close() error {
	zr.ReadCloser.Close()
	return os.Remove(path.Join(zr.path, zr.name))
}

const (
	TempPattern = "temphia_fs_*.zip"
)

func ZipFsFolder(folder fs.FS, slug string) (string, error) {

	file, err := os.CreateTemp(os.TempDir(), TempPattern)
	if err != nil {
		return "", err
	}

	writer := zip.NewWriter(file)
	err = fs.WalkDir(folder, slug, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		file, err := folder.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		zipPath, err := filepath.Rel(slug, filepath.ToSlash(path))
		if err != nil {
			return err
		}

		zipFile, err := writer.Create(zipPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(zipFile, file)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		os.Remove(path.Join(os.TempDir(), file.Name()))
		return "", err
	}

	err = writer.Close()
	if err != nil {
		os.Remove(path.Join(os.TempDir(), file.Name()))
		return "", err
	}

	return file.Name(), nil
}
