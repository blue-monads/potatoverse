package buddyfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/spf13/afero"
)

var (
	_ afero.Fs   = (*BuddyFsClient)(nil)
	_ afero.File = (*OpenFileHandle)(nil)
)

type BuddyFsClient struct {
	httpClient *http.Client
	baseUrl    string
}

func NewBuddyFsClient(baseUrl string) *BuddyFsClient {
	return &BuddyFsClient{
		httpClient: &http.Client{},
		baseUrl:    baseUrl,
	}
}

// Name returns the filesystem name, satisfying afero.Fs.
func (c *BuddyFsClient) Name() string {
	return "buddyfs-http"
}

// ErrorResponse represents an error response from the server
type ErrorResponse struct {
	Message string `json:"message"`
}

// BuddyFsFileInfo implements os.FileInfo.
func (i *BuddyFsFileInfo) Name() string {
	if i == nil {
		return ""
	}
	return i.BaseName
}

func (i *BuddyFsFileInfo) Size() int64 {
	if i == nil {
		return 0
	}
	return i.FileSize
}

func (i *BuddyFsFileInfo) Mode() os.FileMode {
	if i == nil {
		return 0
	}
	return os.FileMode(i.FileMode)
}

func (i *BuddyFsFileInfo) ModTime() time.Time {
	if i == nil {
		return time.Time{}
	}
	return i.Modified
}

func (i *BuddyFsFileInfo) IsDir() bool {
	if i == nil {
		return false
	}
	return i.Directory
}

func (i *BuddyFsFileInfo) Sys() interface{} {
	return nil
}

// File system operations

// Create creates a new file and returns a file handle

func (c *BuddyFsClient) Ping() error {
	req, err := http.NewRequest("POST", c.baseUrl+"/ping", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}
	return nil
}

func (c *BuddyFsClient) Create(name string) (afero.File, error) {

	req, err := http.NewRequest("POST", c.baseUrl+"/create", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("name", name)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &OpenFileHandle{
		parent:  c,
		fileKey: string(body),
		name:    name,
	}, nil
}

// Mkdir creates a directory
func (c *BuddyFsClient) Mkdir(name string, perm os.FileMode) error {
	req, err := http.NewRequest("POST", c.baseUrl+"/mkdir", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("name", name)
	if perm != 0 {
		q.Add("perm", fmt.Sprintf("%04o", perm))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

// MkdirAll creates directories recursively
func (c *BuddyFsClient) MkdirAll(name string, perm os.FileMode) error {
	req, err := http.NewRequest("POST", c.baseUrl+"/mkdirall", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("name", name)
	if perm != 0 {
		q.Add("perm", fmt.Sprintf("%04o", perm))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

// Open opens a file for reading and returns a file handle
func (c *BuddyFsClient) Open(name string) (afero.File, error) {
	req, err := http.NewRequest("POST", c.baseUrl+"/open", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("name", name)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &OpenFileHandle{
		parent:  c,
		fileKey: string(body),
		name:    name,
	}, nil
}

type OpenFileHandle struct {
	parent  *BuddyFsClient
	fileKey string
	name    string
}

// OpenFile opens a file with specified flags and permissions, returns a file handle
func (c *BuddyFsClient) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	req, err := http.NewRequest("POST", c.baseUrl+"/openfile", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("name", name)
	q.Add("flag", strconv.Itoa(flag))
	if perm != 0 {
		q.Add("perm", fmt.Sprintf("%04o", perm))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &OpenFileHandle{
		parent:  c,
		fileKey: string(body),
		name:    name,
	}, nil
}

// Remove removes a file or directory
func (c *BuddyFsClient) Remove(name string) error {
	req, err := http.NewRequest("DELETE", c.baseUrl+"/remove", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("name", name)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

// RemoveAll removes a file or directory recursively
func (c *BuddyFsClient) RemoveAll(name string) error {
	req, err := http.NewRequest("DELETE", c.baseUrl+"/removeall", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("name", name)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

// Rename renames a file or directory.
func (c *BuddyFsClient) Rename(oldname, newname string) error {
	req, err := http.NewRequest("POST", c.baseUrl+"/rename", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("oldname", oldname)
	q.Add("newname", newname)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

// Stat returns file information
func (c *BuddyFsClient) Stat(name string) (os.FileInfo, error) {
	req, err := http.NewRequest("GET", c.baseUrl+"/stat", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("name", name)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var info BuddyFsFileInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}

// Chmod changes file mode
func (c *BuddyFsClient) Chmod(name string, mode os.FileMode) error {
	req, err := http.NewRequest("PUT", c.baseUrl+"/chmod", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("name", name)
	q.Add("mode", fmt.Sprintf("%04o", mode))
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

// Chown changes file ownership
func (c *BuddyFsClient) Chown(name string, uid, gid int) error {
	req, err := http.NewRequest("PUT", c.baseUrl+"/chown", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("name", name)
	q.Add("uid", strconv.Itoa(uid))
	q.Add("gid", strconv.Itoa(gid))
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

// Chtimes changes file access and modification times
func (c *BuddyFsClient) Chtimes(name string, atime, mtime time.Time) error {
	req, err := http.NewRequest("PUT", c.baseUrl+"/chtimes", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("name", name)
	q.Add("atime", atime.Format(time.RFC3339))
	q.Add("mtime", mtime.Format(time.RFC3339))
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

// File operations (methods on OpenFileHandle)

// ensureOpen validates that the file handle is still valid.
func (h *OpenFileHandle) ensureOpen() error {
	if h == nil || h.parent == nil {
		return afero.ErrFileClosed
	}
	if h.fileKey == "" {
		return afero.ErrFileClosed
	}
	return nil
}

// Name returns the cached name of the file, falling back to the server if unknown.
func (h *OpenFileHandle) Name() string {
	if h == nil {
		return ""
	}
	if h.name != "" {
		return h.name
	}

	// Fallback to server lookup. We deliberately ignore errors to comply with the afero.File contract.
	req, err := http.NewRequest("GET", h.parent.baseUrl+"/file/name", nil)
	if err != nil {
		return ""
	}

	q := req.URL.Query()
	q.Add("filekey", h.fileKey)
	req.URL.RawQuery = q.Encode()

	resp, err := h.parent.httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var result struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}

	h.name = result.Name
	return h.name
}

// Readdir reads directory entries.
func (h *OpenFileHandle) Readdir(count int) ([]os.FileInfo, error) {
	if err := h.ensureOpen(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", h.parent.baseUrl+"/file/readdir", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("filekey", h.fileKey)
	if count > 0 {
		q.Add("count", strconv.Itoa(count))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := h.parent.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, h.parent.parseError(resp)
	}

	var files []BuddyFsFileInfo
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	result := make([]os.FileInfo, len(files))
	for i := range files {
		result[i] = &files[i]
	}

	return result, nil
}

// Readdirnames reads directory entry names.
func (h *OpenFileHandle) Readdirnames(n int) ([]string, error) {
	if err := h.ensureOpen(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", h.parent.baseUrl+"/file/readdirnames", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("filekey", h.fileKey)
	if n > 0 {
		q.Add("count", strconv.Itoa(n))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := h.parent.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, h.parent.parseError(resp)
	}

	var names []string
	if err := json.NewDecoder(resp.Body).Decode(&names); err != nil {
		return nil, err
	}

	return names, nil
}

// Stat returns file information for an open file.
func (h *OpenFileHandle) Stat() (os.FileInfo, error) {
	if err := h.ensureOpen(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", h.parent.baseUrl+"/file/stat", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("filekey", h.fileKey)
	req.URL.RawQuery = q.Encode()

	resp, err := h.parent.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, h.parent.parseError(resp)
	}

	var info BuddyFsFileInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}

// Sync synchronizes the file
func (h *OpenFileHandle) Sync() error {
	if err := h.ensureOpen(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", h.parent.baseUrl+"/file/sync", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("filekey", h.fileKey)
	req.URL.RawQuery = q.Encode()

	resp, err := h.parent.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return h.parent.parseError(resp)
	}

	return nil
}

// Truncate truncates the file to the specified size
func (h *OpenFileHandle) Truncate(size int64) error {
	if err := h.ensureOpen(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", h.parent.baseUrl+"/file/truncate", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("filekey", h.fileKey)
	q.Add("size", strconv.FormatInt(size, 10))
	req.URL.RawQuery = q.Encode()

	resp, err := h.parent.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return h.parent.parseError(resp)
	}

	return nil
}

// WriteString writes data from a string to the file.
func (h *OpenFileHandle) WriteString(data string) (int, error) {
	return h.Write([]byte(data))
}

// Write writes data from the provided slice to the file.
func (h *OpenFileHandle) Write(p []byte) (int, error) {
	if err := h.ensureOpen(); err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", h.parent.baseUrl+"/file/write", bytes.NewReader(p))
	if err != nil {
		return 0, err
	}

	q := req.URL.Query()
	q.Add("filekey", h.fileKey)
	req.URL.RawQuery = q.Encode()

	resp, err := h.parent.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, h.parent.parseError(resp)
	}

	var result struct {
		Written int64 `json:"written"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	written := int(result.Written)
	if int64(written) != result.Written {
		return 0, fmt.Errorf("written bytes exceed int range")
	}

	return written, nil
}

// WriteAt writes data to the file at the specified offset.
func (h *OpenFileHandle) WriteAt(p []byte, off int64) (int, error) {
	if err := h.ensureOpen(); err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", h.parent.baseUrl+"/file/writeat", bytes.NewReader(p))
	if err != nil {
		return 0, err
	}

	q := req.URL.Query()
	q.Add("filekey", h.fileKey)
	q.Add("off", strconv.FormatInt(off, 10))
	req.URL.RawQuery = q.Encode()

	resp, err := h.parent.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, h.parent.parseError(resp)
	}

	var result struct {
		Written int64 `json:"written"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	written := int(result.Written)
	if int64(written) != result.Written {
		return 0, fmt.Errorf("written bytes exceed int range")
	}

	return written, nil
}

// Read reads data from the server-backed file into p.
func (h *OpenFileHandle) Read(p []byte) (int, error) {
	if err := h.ensureOpen(); err != nil {
		return 0, err
	}
	if len(p) == 0 {
		return 0, nil
	}

	req, err := http.NewRequest("POST", h.parent.baseUrl+"/file/read", nil)
	if err != nil {
		return 0, err
	}

	q := req.URL.Query()
	q.Add("filekey", h.fileKey)
	q.Add("size", strconv.Itoa(len(p)))
	req.URL.RawQuery = q.Encode()

	resp, err := h.parent.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, h.parent.parseError(resp)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// Attempt to decode JSON payloads (legacy behaviour) before assuming raw bytes.
	if json.Valid(data) {
		var payload struct {
			Data []byte `json:"data"`
			Read int    `json:"read"`
		}
		if err := json.Unmarshal(data, &payload); err == nil {
			if len(payload.Data) > 0 {
				n := copy(p, payload.Data)
				if n < len(payload.Data) {
					return n, io.ErrShortBuffer
				}
				if payload.Read > n {
					return n, io.ErrShortBuffer
				}
				if payload.Read < n {
					return payload.Read, io.EOF
				}
				return n, nil
			}
			if payload.Read == 0 {
				return 0, io.EOF
			}
			if payload.Read > len(p) {
				return len(p), io.ErrShortBuffer
			}
			return payload.Read, io.EOF
		}
	}

	if len(data) == 0 {
		return 0, io.EOF
	}

	if len(data) > len(p) {
		copy(p, data[:len(p)])
		return len(p), io.ErrShortBuffer
	}

	n := copy(p, data)
	if n < len(p) {
		return n, io.EOF
	}

	return n, nil
}

// ReadAt reads data from the file at a specific offset into p.
func (h *OpenFileHandle) ReadAt(p []byte, off int64) (int, error) {
	if err := h.ensureOpen(); err != nil {
		return 0, err
	}
	if len(p) == 0 {
		return 0, nil
	}

	req, err := http.NewRequest("POST", h.parent.baseUrl+"/file/readat", nil)
	if err != nil {
		return 0, err
	}

	q := req.URL.Query()
	q.Add("filekey", h.fileKey)
	q.Add("off", strconv.FormatInt(off, 10))
	q.Add("size", strconv.Itoa(len(p)))
	req.URL.RawQuery = q.Encode()

	resp, err := h.parent.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, h.parent.parseError(resp)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if json.Valid(data) {
		var payload struct {
			Data []byte `json:"data"`
			Read int    `json:"read"`
		}
		if err := json.Unmarshal(data, &payload); err == nil {
			if len(payload.Data) > 0 {
				n := copy(p, payload.Data)
				if n < len(payload.Data) {
					return n, io.ErrShortBuffer
				}
				if payload.Read > n {
					return n, io.ErrShortBuffer
				}
				if payload.Read < n {
					return payload.Read, io.EOF
				}
				return n, nil
			}
			if payload.Read == 0 {
				return 0, io.EOF
			}
			if payload.Read > len(p) {
				return len(p), io.ErrShortBuffer
			}
			return payload.Read, io.EOF
		}
	}

	if len(data) == 0 {
		return 0, io.EOF
	}

	if len(data) > len(p) {
		copy(p, data[:len(p)])
		return len(p), io.ErrShortBuffer
	}

	n := copy(p, data)
	if n < len(p) {
		return n, io.EOF
	}

	return n, nil
}

// Seek seeks to a specific offset in the file
// Note: Server only supports io.SeekStart (whence = 0)
func (h *OpenFileHandle) Seek(offset int64, whence int) (int64, error) {
	if err := h.ensureOpen(); err != nil {
		return 0, err
	}

	if whence != 0 { // io.SeekStart
		// Server only supports SeekStart, so we need to calculate offset for other whence values
		if whence == 1 { // io.SeekCurrent
			// Get current position by seeking to 0 and getting result, then add offset
			// But we can't easily get current position, so return error
			return 0, fmt.Errorf("SeekCurrent not supported by server")
		}
		if whence == 2 { // io.SeekEnd
			// Would need file size, return error
			return 0, fmt.Errorf("SeekEnd not supported by server")
		}
		return 0, fmt.Errorf("invalid whence value: %d", whence)
	}

	req, err := http.NewRequest("POST", h.parent.baseUrl+"/file/seek", nil)
	if err != nil {
		return 0, err
	}

	q := req.URL.Query()
	q.Add("filekey", h.fileKey)
	q.Add("off", strconv.FormatInt(offset, 10))
	req.URL.RawQuery = q.Encode()

	resp, err := h.parent.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, h.parent.parseError(resp)
	}

	var result struct {
		Seeked int64 `json:"seeked"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.Seeked, nil
}

// Close closes the file
func (h *OpenFileHandle) Close() error {
	if err := h.ensureOpen(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", h.parent.baseUrl+"/file/close", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("filekey", h.fileKey)
	req.URL.RawQuery = q.Encode()

	resp, err := h.parent.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return h.parent.parseError(resp)
	}

	h.fileKey = ""

	return nil
}

// parseError parses an error response from the server
func (c *BuddyFsClient) parseError(resp *http.Response) error {
	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, resp.Status)
	}
	return fmt.Errorf("%s", errResp.Message)
}
