package xutils

import (
	"io"
	"os"
	"path"
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
