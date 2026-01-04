package file

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
)

// AsFS implements fs.FS interface.
//
// Example usage with html/template:
//
//	fileFS := NewAsFS(fileOps, ownerID, "templates")
//	tmpl, err := template.ParseFS(fileFS, "index.html")
//
// Example usage with fs.ReadFile:
//
//	fileFS := NewAsFS(fileOps, ownerID, "")
//	data, err := fs.ReadFile(fileFS, "templates/index.html")
type AsFS struct {
	ops      *FileOperations
	rootPath string
	ownerID  int64
}

// NewAsFS creates a new AsFS instance
func (o *FileOperations) NewAsFS(ownerID int64, rootPath string) fs.FS {
	return &AsFS{
		ops:      o,
		rootPath: rootPath,
		ownerID:  ownerID,
	}
}

// Open opens the named file for reading
func (a *AsFS) Open(name string) (fs.File, error) {
	// Clean and validate the path
	name = filepath.Clean(name)
	if !fs.ValidPath(name) {
		return nil, fmt.Errorf("open %s: %w", name, fs.ErrInvalid)
	}

	// Split path into directory and filename
	dir, filename := filepath.Split(name)
	dir = strings.TrimSuffix(dir, "/")

	// Combine root path with the requested path
	fullPath := a.rootPath
	if dir != "" && dir != "." {
		if fullPath == "" {
			fullPath = dir
		} else {
			fullPath = filepath.Join(fullPath, dir)
		}
	}

	// Get file metadata
	fileMeta, err := a.ops.GetFileMetaByPath(a.ownerID, fullPath, filename)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", name, fs.ErrNotExist)
	}

	// Handle directories
	if fileMeta.IsFolder {
		return newDirFile(a, name, fileMeta), nil
	}

	// Handle regular files
	return newFile(a, name, fileMeta)
}

// file implements fs.File for regular files
type file struct {
	fs      *AsFS
	name    string
	meta    *dbmodels.FileMeta
	content []byte
	reader  *bytes.Reader
}

func newFile(fs *AsFS, name string, meta *dbmodels.FileMeta) (*file, error) {
	// Read file content
	content, err := fs.ops.getFileContentByMeta(meta)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}

	return &file{
		fs:      fs,
		name:    name,
		meta:    meta,
		content: content,
		reader:  bytes.NewReader(content),
	}, nil
}

func (f *file) Stat() (fs.FileInfo, error) {
	return &fileInfo{f.meta}, nil
}

func (f *file) Read(p []byte) (int, error) {
	return f.reader.Read(p)
}

func (f *file) Close() error {
	f.reader = bytes.NewReader(f.content) // Reset for potential reuse
	return nil
}

// dirFile implements fs.File and fs.ReadDirFile for directories
type dirFile struct {
	fs      *AsFS
	name    string
	meta    *dbmodels.FileMeta
	entries []*dirEntry
	pos     int
}

func newDirFile(fs *AsFS, name string, meta *dbmodels.FileMeta) *dirFile {
	// Build the full path for listing
	fullPath := fs.rootPath
	if name != "." && name != "" {
		// Remove the root path prefix if present
		if strings.HasPrefix(name, fs.rootPath) {
			fullPath = strings.TrimPrefix(name, fs.rootPath)
			fullPath = strings.TrimPrefix(fullPath, "/")
		} else {
			fullPath = name
		}
	}

	// List files in the directory
	files, err := fs.ops.ListFiles(fs.ownerID, fullPath)
	if err != nil {
		files = []dbmodels.FileMeta{} // Return empty on error
	}

	entries := make([]*dirEntry, 0, len(files))
	for i := range files {
		entries = append(entries, &dirEntry{&files[i]})
	}

	return &dirFile{
		fs:      fs,
		name:    name,
		meta:    meta,
		entries: entries,
		pos:     0,
	}
}

func (d *dirFile) Stat() (fs.FileInfo, error) {
	return &fileInfo{d.meta}, nil
}

func (d *dirFile) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("read %s: is a directory", d.name)
}

func (d *dirFile) Close() error {
	d.pos = 0
	return nil
}

func (d *dirFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.pos >= len(d.entries) {
		if n <= 0 {
			return nil, nil
		}
		return nil, io.EOF
	}

	if n <= 0 {
		// Return all remaining entries
		entries := make([]fs.DirEntry, len(d.entries)-d.pos)
		for i, e := range d.entries[d.pos:] {
			entries[i] = e
		}
		d.pos = len(d.entries)
		return entries, nil
	}

	// Return at most n entries
	end := d.pos + n
	if end > len(d.entries) {
		end = len(d.entries)
	}
	entries := make([]fs.DirEntry, end-d.pos)
	for i, e := range d.entries[d.pos:end] {
		entries[i] = e
	}
	d.pos = end

	if d.pos >= len(d.entries) && n > 0 {
		return entries, io.EOF
	}

	return entries, nil
}

// fileInfo implements fs.FileInfo
type fileInfo struct {
	meta *dbmodels.FileMeta
}

func (fi *fileInfo) Name() string {
	return fi.meta.Name
}

func (fi *fileInfo) Size() int64 {
	return fi.meta.Size
}

func (fi *fileInfo) Mode() fs.FileMode {
	if fi.meta.IsFolder {
		return fs.ModeDir | 0755
	}
	return 0644
}

func (fi *fileInfo) ModTime() time.Time {
	if fi.meta.UpdatedAt != nil {
		return *fi.meta.UpdatedAt
	}
	if fi.meta.CreatedAt != nil {
		return *fi.meta.CreatedAt
	}
	return time.Time{}
}

func (fi *fileInfo) IsDir() bool {
	return fi.meta.IsFolder
}

func (fi *fileInfo) Sys() interface{} {
	return fi.meta
}

// dirEntry implements fs.DirEntry
type dirEntry struct {
	meta *dbmodels.FileMeta
}

func (de *dirEntry) Name() string {
	return de.meta.Name
}

func (de *dirEntry) IsDir() bool {
	return de.meta.IsFolder
}

func (de *dirEntry) Type() fs.FileMode {
	if de.meta.IsFolder {
		return fs.ModeDir
	}
	return 0
}

func (de *dirEntry) Info() (fs.FileInfo, error) {
	return &fileInfo{de.meta}, nil
}
