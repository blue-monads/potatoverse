package buddyfs

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	xutils "github.com/blue-monads/turnix/backend/utils"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
)

type BuddyFsFile struct {
	afero.File
	lastAccessed time.Time
}

type BuddyFs struct {
	fs     *os.Root
	fsPath string

	files map[string]*BuddyFsFile
	mu    sync.Mutex
}

func (b *BuddyFs) Create(basePath string, ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		httpx.WriteErr(ctx, errors.New("name is required"))
		return
	}

	file, err := b.fs.Create(filepath.Join(basePath, name))
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	randStr, err := xutils.GenerateRandomString(10)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	f := &BuddyFsFile{
		File:         file,
		lastAccessed: time.Now(),
	}

	b.mu.Lock()
	b.files[randStr] = f
	b.mu.Unlock()

	ctx.Data(http.StatusOK, "", []byte(randStr))
}

func (b *BuddyFs) Mkdir(basePath string, ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		httpx.WriteErr(ctx, errors.New("name is required"))
		return
	}

	perm := ctx.Query("perm")
	if perm == "" {
		perm = "0755"
	}

	permInt, err := strconv.ParseInt(perm, 8, 64)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	err = b.fs.Mkdir(filepath.Join(basePath, name), os.FileMode(permInt))
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteOk(ctx)
}

func (b *BuddyFs) MkdirAll(basePath string, ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		httpx.WriteErr(ctx, errors.New("name is required"))
		return
	}

	perm := ctx.Query("perm")
	if perm == "" {
		perm = "0755"
	}

	permInt, err := strconv.ParseInt(perm, 8, 64)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	splitPath := strings.Split(filepath.Join(basePath, name), "/")

	for _, path := range splitPath {
		if path == "" {
			continue
		}
		b.fs.Mkdir(path, os.FileMode(permInt))
	}

	httpx.WriteOk(ctx)

}

func (b *BuddyFs) Open(basePath string, ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		httpx.WriteErr(ctx, errors.New("name is required"))
		return
	}

	file, err := b.fs.Open(filepath.Join(basePath, name))
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	randStr, err := xutils.GenerateRandomString(10)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	f := &BuddyFsFile{
		File:         file,
		lastAccessed: time.Now(),
	}

	b.mu.Lock()
	b.files[randStr] = f
	b.mu.Unlock()
}

func (b *BuddyFs) OpenFile(basePath string, ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		httpx.WriteErr(ctx, errors.New("name is required"))
		return
	}

	flag := ctx.Query("flag")
	if flag == "" {
		flag = "0"
	}

	flagInt, err := strconv.ParseInt(flag, 10, 64)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
	perm := ctx.Query("perm")
	if perm == "" {
		perm = "0644"
	}

	permInt, err := strconv.ParseInt(perm, 8, 64)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	file, err := b.fs.OpenFile(filepath.Join(basePath, name), int(flagInt), os.FileMode(permInt))
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	randStr, err := xutils.GenerateRandomString(10)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	f := &BuddyFsFile{
		File:         file,
		lastAccessed: time.Now(),
	}

	b.mu.Lock()
	b.files[randStr] = f
	b.mu.Unlock()
}

func (b *BuddyFs) Remove(basePath string, ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		httpx.WriteErr(ctx, errors.New("name is required"))
		return
	}

	err := b.fs.Remove(filepath.Join(basePath, name))
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
	httpx.WriteOk(ctx)
}

func (b *BuddyFs) RemoveAll(basePath string, ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		httpx.WriteErr(ctx, errors.New("name is required"))
		return
	}

	err := b.fs.Remove(filepath.Join(basePath, name))
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
	httpx.WriteOk(ctx)
}

func (b *BuddyFs) Rename(basePath string, ctx *gin.Context) {
	panic("not implemented")
}

type BuddyFsFileInfo struct {
	Name    string
	Size    int64
	Mode    int64
	ModTime time.Time
	IsDir   bool
}

func (b *BuddyFs) Stat(basePath string, ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		httpx.WriteErr(ctx, errors.New("name is required"))
		return
	}

	fi, err := b.fs.Stat(name)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	bfi := &BuddyFsFileInfo{
		Name:    fi.Name(),
		Size:    fi.Size(),
		Mode:    int64(fi.Mode()),
		ModTime: fi.ModTime(),
		IsDir:   fi.IsDir(),
	}

	httpx.WriteJSON(ctx, bfi, nil)
}

func (b *BuddyFs) Chmod(basePath string, ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		httpx.WriteErr(ctx, errors.New("name is required"))
		return
	}

	mode := ctx.Query("mode")
	if mode == "" {
		mode = "0644"
	}

	modeInt, err := strconv.ParseInt(mode, 8, 64)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	fi, err := b.fs.Open(filepath.Join(basePath, name))
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
	defer fi.Close()

	err = fi.Chmod(os.FileMode(modeInt))
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteOk(ctx)
}

func (b *BuddyFs) Chown(basePath string, ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		httpx.WriteErr(ctx, errors.New("name is required"))
		return
	}

	uidStr := ctx.Query("uid")
	gidStr := ctx.Query("gid")

	uidInt, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
	gidInt, err := strconv.ParseInt(gidStr, 10, 64)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	fi, err := b.fs.Open(filepath.Join(basePath, name))
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
	defer fi.Close()

	err = fi.Chown(int(uidInt), int(gidInt))
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteOk(ctx)
}

func (b *BuddyFs) Chtimes(basePath string, ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		httpx.WriteErr(ctx, errors.New("name is required"))
		return
	}

	atime := ctx.Query("atime")
	if atime == "" {
		atime = time.Now().Format(time.RFC3339)
	}

	mtime := ctx.Query("mtime")
	if mtime == "" {
		mtime = time.Now().Format(time.RFC3339)
	}

	atimeTime, err := time.Parse(time.RFC3339, atime)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	mtimeTime, err := time.Parse(time.RFC3339, mtime)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	err = os.Chtimes(filepath.Join(b.fsPath, basePath, name), atimeTime, mtimeTime)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteOk(ctx)
}
