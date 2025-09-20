package server

import (
	"fmt"
	"strconv"

	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/gin-gonic/gin"
)

func (a *Server) ListSpaceFiles(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	spaceId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this space
	space, err := a.ctrl.GetSpace(spaceId)
	if err != nil {
		return nil, err
	}

	if space.OwnerID != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this space")
	}

	path := ctx.Query("path")

	files, err := a.ctrl.ListSpaceFiles(spaceId, path)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (a *Server) GetSpaceFile(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	spaceId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	fileId, err := strconv.ParseInt(ctx.Param("fileId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this space
	space, err := a.ctrl.GetSpace(spaceId)
	if err != nil {
		return nil, err
	}

	if space.OwnerID != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this space")
	}

	file, err := a.ctrl.GetSpaceFile(spaceId, fileId)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (a *Server) DownloadSpaceFile(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	spaceId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	fileId, err := strconv.ParseInt(ctx.Param("fileId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this space
	space, err := a.ctrl.GetSpace(spaceId)
	if err != nil {
		return nil, err
	}

	if space.OwnerID != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this space")
	}

	// Get file metadata first
	file, err := a.ctrl.GetSpaceFile(spaceId, fileId)
	if err != nil {
		return nil, err
	}

	// Set headers for file download
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.Name))
	ctx.Header("Content-Length", fmt.Sprintf("%d", file.Size))

	// Stream the file content
	err = a.ctrl.DownloadSpaceFile(spaceId, fileId, ctx.Writer)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (a *Server) DeleteSpaceFile(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	spaceId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	fileId, err := strconv.ParseInt(ctx.Param("fileId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this space
	space, err := a.ctrl.GetSpace(spaceId)
	if err != nil {
		return nil, err
	}

	if space.OwnerID != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this space")
	}

	err = a.ctrl.DeleteSpaceFile(spaceId, fileId)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "File deleted successfully"}, nil
}

func (a *Server) UploadSpaceFile(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	spaceId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this space
	space, err := a.ctrl.GetSpace(spaceId)
	if err != nil {
		return nil, err
	}

	if space.OwnerID != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this space")
	}

	// Parse multipart form
	err = ctx.Request.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		return nil, err
	}

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	path := ctx.Request.FormValue("path")

	fileId, err := a.ctrl.UploadSpaceFile(spaceId, header.Filename, path, file, claim.UserId)
	if err != nil {
		return nil, err
	}

	return gin.H{"file_id": fileId, "message": "File uploaded successfully"}, nil
}

func (a *Server) CreateSpaceFolder(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	spaceId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this space
	space, err := a.ctrl.GetSpace(spaceId)
	if err != nil {
		return nil, err
	}

	if space.OwnerID != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this space")
	}

	// Parse request body
	var req struct {
		Name string `json:"name" binding:"required"`
		Path string `json:"path"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	folderId, err := a.ctrl.CreateSpaceFolder(spaceId, req.Name, req.Path, claim.UserId)
	if err != nil {
		return nil, err
	}

	return gin.H{"folder_id": folderId, "message": "Folder created successfully"}, nil
}
