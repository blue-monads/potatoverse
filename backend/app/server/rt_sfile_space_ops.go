package server

import (
	"fmt"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/corehub"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/utils/libx/httpx"
	"github.com/gin-gonic/gin"
)

func (s *Server) parseSpaceToken(ctx *gin.Context) (*signer.SpaceClaim, error) {
	tok := ctx.GetHeader("Autorization")

	// fixme =>  check permission

	claim, err := s.signer.ParseSpace(tok)
	if err != nil {
		return nil, err
	}

	return claim, nil
}

func (s *Server) spaceFileList(ctx *gin.Context) {
	claim, err := s.parseSpaceToken(ctx)
	if err != nil {
		httpx.UnAuthorized(ctx)
		return
	}

	path := ctx.Query("path")

	files, err := s.opt.CoreHub.ListSpaceFilesSigned(claim.InstallId, path)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteJSON(ctx, files, nil)
}

func (s *Server) spaceFileDownload(ctx *gin.Context) {

	refId := ctx.Param("ref_id")

	s.opt.CoreHub.ServeSpaceFileSigned(refId, ctx)
}

func (s *Server) spaceFilePreview(ctx *gin.Context) {
	refId := ctx.Param("ref_id")

	s.opt.CoreHub.ServePreviewFileSigned(refId, ctx)
}

func (s *Server) spaceFileUpload(ctx *gin.Context) {

	claim, err := s.parseSpaceToken(ctx)
	if err != nil {
		httpx.UnAuthorized(ctx)
		return
	}

	err = ctx.Request.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	file, _, err := ctx.Request.FormFile("files")
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}
	defer file.Close()

	fileName := ctx.Request.Form.Get("filename")

	if fileName == "" {
		httpx.WriteErr(ctx, fmt.Errorf("filename is required"))
		return
	}

	fpath := ctx.Query("path")
	fpath = strings.TrimPrefix(fpath, "/")
	fpath = strings.TrimSuffix(fpath, "/")

	fileId, err := s.ctrl.UploadSpaceFile(claim.InstallId, fileName, fpath, file, claim.UserId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": fmt.Sprintf("Failed to upload file: %v", err)})
		return
	}

	fileSignedId, err := corehub.SignFileId(s.signer, fileId)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	ctx.JSON(200, gin.H{
		"file_id": fileSignedId,
		"message": "File uploaded successfully",
	})

}

func (s *Server) spaceFileCreateFolder(ctx *gin.Context) {
	claim, err := s.parseSpaceToken(ctx)
	if err != nil {
		httpx.UnAuthorized(ctx)
		return
	}

	var req CreateSpaceFolderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	folderId, err := s.ctrl.CreateSpaceFolder(claim.InstallId, req.Name, req.Path, claim.UserId)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	folderSignedId, err := corehub.SignFileId(s.signer, folderId)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	ctx.JSON(200, gin.H{
		"folder_id": folderSignedId,
		"message":   "Folder created successfully",
	})
}
