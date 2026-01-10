package corehub

import (
	"net/http"

	"github.com/blue-monads/potatoverse/backend/utils/libx/httpx"
	"github.com/gin-gonic/gin"
)

type FileMeta struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	IsFolder bool   `json:"is_folder"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Mime     string `json:"mime"`
	Hash     string `json:"hash"`
}

const (
	Salt = "signed-file"
)

func (c *CoreHub) ListSpaceFilesSigned(installId int64, path string) ([]FileMeta, error) {
	files, err := c.db.GetFileOps().ListFiles(installId, path)
	if err != nil {
		return nil, err
	}

	ffiles := make([]FileMeta, len(files))
	refIds := make([]string, len(files))

	for i, file := range files {
		refIds[i] = file.RefID
	}

	altSignedRefIds, err := c.signer.SignAltBatch(Salt, refIds)
	if err != nil {
		return nil, err
	}

	for i, file := range files {
		ffiles[i] = FileMeta{
			Id:       altSignedRefIds[i],
			Name:     file.Name,
			IsFolder: file.IsFolder,
			Path:     file.Path,
			Size:     file.Size,
			Mime:     file.Mime,
			Hash:     file.Hash,
		}
	}

	return ffiles, nil
}

func (c *CoreHub) ServeSpaceFileSigned(refId string, ctx *gin.Context) {
	originalRefId, _, err := c.signer.VerifyAlt(Salt, refId)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	fid, err := c.db.GetFileOps().GetFileByRefId(originalRefId)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	err = c.db.GetFileOps().StreamFileToHTTP(fid.OwnerID, fid.Path, fid.Name, ctx)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
}

func (c *CoreHub) ServePreviewFileSigned(refId string, ctx *gin.Context) {
	originalRefId, _, err := c.signer.VerifyAlt(Salt, refId)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	fid, err := c.db.GetFileOps().GetFileByRefId(originalRefId)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	preview, err := c.db.GetFileOps().GetFilePreview(fid.OwnerID, fid.ID)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	ctx.Data(http.StatusOK, "image/jpeg", preview)
}
