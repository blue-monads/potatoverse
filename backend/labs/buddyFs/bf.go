package buddyfs

import (
	"errors"
	"io"
	"strconv"

	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/gin-gonic/gin"
)

func (b *BuddyFs) getFile(ctx *gin.Context) (*BuddyFsFile, error) {
	fileKey := ctx.Query("filekey")
	if fileKey == "" {
		httpx.WriteErr(ctx, errors.New("filekey is required"))
		return nil, errors.New("filekey is required")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	file, ok := b.files[fileKey]
	if !ok {
		return nil, errors.New("file not found")
	}

	return file, nil
}

func (b *BuddyFs) Name(ctx *gin.Context) {

	file, err := b.getFile(ctx)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteJSON(ctx,
		map[string]string{
			"name": file.Name(),
		}, nil)

}

func (b *BuddyFs) Readdir(ctx *gin.Context) {
	file, err := b.getFile(ctx)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	count := ctx.Query("count")
	if count == "" {
		count = "10"
	}

	countInt, err := strconv.Atoi(count)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	files, err := file.Readdir(countInt)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	rFiles := make([]BuddyFsFileInfo, len(files))
	for i, file := range files {
		rFiles[i] = BuddyFsFileInfo{
			Name:    file.Name(),
			Size:    file.Size(),
			Mode:    int64(file.Mode()),
			ModTime: file.ModTime(),
			IsDir:   file.IsDir(),
		}
	}

	httpx.WriteJSON(ctx, rFiles, nil)
}

func (b *BuddyFs) Readdirnames(ctx *gin.Context) {
	file, err := b.getFile(ctx)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	names, err := file.Readdirnames(10)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteJSON(ctx, names, nil)
}

func (b *BuddyFs) FileStat(ctx *gin.Context) {
	file, err := b.getFile(ctx)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	fi, err := file.Stat()
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteJSON(ctx, BuddyFsFileInfo{
		Name:    fi.Name(),
		Size:    fi.Size(),
		Mode:    int64(fi.Mode()),
		ModTime: fi.ModTime(),
		IsDir:   fi.IsDir(),
	}, nil)
}

func (b *BuddyFs) Sync(ctx *gin.Context) {
	file, err := b.getFile(ctx)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	err = file.Sync()
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
	httpx.WriteOk(ctx)
}

func (b *BuddyFs) Truncate(ctx *gin.Context) {
	file, err := b.getFile(ctx)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	size := ctx.Query("size")
	if size == "" {
		httpx.WriteErr(ctx, errors.New("size is required"))
		return
	}

	sizeInt, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	err = file.Truncate(sizeInt)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteOk(ctx)
}

func (b *BuddyFs) WriteString(ctx *gin.Context) {
	file, err := b.getFile(ctx)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	written, err := io.Copy(file, ctx.Request.Body)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteJSON(ctx, map[string]int64{"written": written}, nil)
}

func (b *BuddyFs) Close(ctx *gin.Context) {
	file, err := b.getFile(ctx)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.files, file.Name())

	err = file.Close()
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
	httpx.WriteOk(ctx)
}

func (b *BuddyFs) Read(ctx *gin.Context) {
	file, err := b.getFile(ctx)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	_, err = file.Write(bodyBytes)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteJSON(ctx, map[string]int{"read": len(bodyBytes)}, nil)
}

func (b *BuddyFs) ReadAt(ctx *gin.Context) {
	file, err := b.getFile(ctx)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	off := ctx.Query("off")
	if off == "" {
		httpx.WriteErr(ctx, errors.New("off is required"))
		return
	}

	offInt, err := strconv.ParseInt(off, 10, 64)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	_, err = file.WriteAt(bodyBytes, offInt)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteJSON(ctx, map[string]int{"read": len(bodyBytes)}, nil)
}

func (b *BuddyFs) Seek(ctx *gin.Context) {
	file, err := b.getFile(ctx)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	off := ctx.Query("off")
	if off == "" {
		httpx.WriteErr(ctx, errors.New("off is required"))
		return
	}

	offInt, err := strconv.ParseInt(off, 10, 64)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	newOff, err := file.Seek(offInt, io.SeekStart)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteJSON(ctx, map[string]int64{"seeked": newOff}, nil)
}

func (b *BuddyFs) Write(ctx *gin.Context) {
	file, err := b.getFile(ctx)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	written, err := io.Copy(file, ctx.Request.Body)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteJSON(ctx, map[string]int64{"written": written}, nil)
}
