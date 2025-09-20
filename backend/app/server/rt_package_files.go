package server

import (
	"fmt"
	"strconv"

	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/gin-gonic/gin"
)

func (a *Server) ListPackageFiles(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	packageId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this package
	pkg, err := a.ctrl.GetPackage(packageId)
	if err != nil {
		return nil, err
	}

	if pkg.InstalledBy != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this package")
	}

	files, err := a.ctrl.ListPackageFiles(packageId)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (a *Server) GetPackageFile(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	packageId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	fileId, err := strconv.ParseInt(ctx.Param("fileId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this package
	pkg, err := a.ctrl.GetPackage(packageId)
	if err != nil {
		return nil, err
	}

	if pkg.InstalledBy != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this package")
	}

	file, err := a.ctrl.GetPackageFile(packageId, fileId)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (a *Server) DownloadPackageFile(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	packageId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	fileId, err := strconv.ParseInt(ctx.Param("fileId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this package
	pkg, err := a.ctrl.GetPackage(packageId)
	if err != nil {
		return nil, err
	}

	if pkg.InstalledBy != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this package")
	}

	// Get file metadata first
	file, err := a.ctrl.GetPackageFile(packageId, fileId)
	if err != nil {
		return nil, err
	}

	// Set headers for file download
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.Name))
	ctx.Header("Content-Length", fmt.Sprintf("%d", file.Size))

	// Stream the file content
	err = a.ctrl.DownloadPackageFile(packageId, fileId, ctx.Writer)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (a *Server) DeletePackageFile(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	packageId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	fileId, err := strconv.ParseInt(ctx.Param("fileId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this package
	pkg, err := a.ctrl.GetPackage(packageId)
	if err != nil {
		return nil, err
	}

	if pkg.InstalledBy != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this package")
	}

	err = a.ctrl.DeletePackageFile(packageId, fileId)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "File deleted successfully"}, nil
}

func (a *Server) UploadPackageFile(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	packageId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this package
	pkg, err := a.ctrl.GetPackage(packageId)
	if err != nil {
		return nil, err
	}

	if pkg.InstalledBy != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this package")
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
	if path == "" {
		path = "/"
	}

	fileId, err := a.ctrl.UploadPackageFile(packageId, header.Filename, path, file)
	if err != nil {
		return nil, err
	}

	return gin.H{"file_id": fileId, "message": "File uploaded successfully"}, nil
}
