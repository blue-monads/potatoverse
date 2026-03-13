package cfiles

import (
	"encoding/base64"
	"errors"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/utils/libx/httpx"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

type FilesCapability struct {
	fileOps   datahub.FileOps
	installId int64
	handle    xcapability.XCapabilityHandle
}

func (c *FilesCapability) Handle(ctx *gin.Context) {
	action := ctx.Param("action")

	claim, err := c.handle.ParseCapToken(ctx.Query("file_token"))
	if err != nil {
		ctx.JSON(401, gin.H{"error": err.Error()})
		return
	}

	if claim.CapabilityId != c.handle.GetModel().ID {
		ctx.JSON(401, gin.H{"error": "invalid capability id"})
		return
	}

	switch action {
	case "stream_file":
		if claim.SubType != "serve_file" {
			ctx.JSON(401, gin.H{"error": "invalid sub type"})
			return
		}

		id, err := strconv.ParseInt(claim.ResourceId, 10, 64)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "invalid resource id"})
			return
		}

		c.handleStreamFile(ctx, id)
	case "stream_file_by_path":
		if claim.SubType != "serve_file" {
			ctx.JSON(401, gin.H{"error": "invalid sub type"})
			return
		}

		dir, file := filepath.Split(claim.ResourceId)
		c.handleStreamFileByPath(ctx, dir, file)
	default:
		if claim.SubType != "operation" {
			ctx.JSON(401, gin.H{"error": "invalid sub type"})
			return
		}

		result, err := c.Execute(action, lazydata.NewLazyHTTP(ctx))
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		httpx.WriteJSON(ctx, result, err)
	}
}

func (c *FilesCapability) ListActions() ([]string, error) {
	return []string{
		"create_file",
		"create_folder",
		"get_file_content",
		"get_file_content_by_path",
		"get_file_content_base64",
		"get_file_meta",
		"get_file_meta_by_path",
		"list_files",
		"remove_file",
		"update_file",
		"update_file_meta",
	}, nil
}

func (c *FilesCapability) Execute(name string, params lazydata.LazyData) (any, error) {
	switch name {
	case "create_file":
		return c.createFile(params)
	case "create_folder":
		return c.createFolder(params)
	case "get_file_content":
		return c.getFileContent(params)
	case "get_file_content_by_path":
		return c.getFileContentByPath(params)
	case "get_file_content_base64":
		return c.getFileContentBase64(params)
	case "get_file_meta":
		return c.getFileMeta(params)
	case "get_file_meta_by_path":
		return c.getFileMetaByPath(params)
	case "list_files":
		return c.listFiles(params)
	case "remove_file":
		return c.removeFile(params)
	case "update_file":
		return c.updateFile(params)
	case "update_file_meta":
		return c.updateFileMeta(params)
	default:
		return nil, errors.New("unknown action: " + name)
	}
}

func (c *FilesCapability) Reload(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {
	return c, nil
}

func (c *FilesCapability) Close() error {
	return nil
}

// action params

type createFileParams struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

type createFolderParams struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type fileIdParams struct {
	ID int64 `json:"id"`
}

type filePathParams struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type updateFileParams struct {
	ID      int64  `json:"id"`
	Content string `json:"content"`
}

type updateFileMetaParams struct {
	ID   int64          `json:"id"`
	Data map[string]any `json:"data"`
}

type listFilesParams struct {
	Path string `json:"path"`
}

// actions

func (c *FilesCapability) createFile(params lazydata.LazyData) (any, error) {
	var p createFileParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	if p.Name == "" {
		return nil, errors.New("name is required")
	}

	fileId, err := c.fileOps.CreateFile(c.installId, &datahub.CreateFileRequest{
		Name:      p.Name,
		Path:      p.Path,
		CreatedBy: c.installId,
	}, strings.NewReader(p.Content))
	if err != nil {
		return nil, err
	}

	return map[string]any{"id": fileId}, nil
}

func (c *FilesCapability) createFolder(params lazydata.LazyData) (any, error) {
	var p createFolderParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	if p.Name == "" {
		return nil, errors.New("name is required")
	}

	folderId, err := c.fileOps.CreateFolder(c.installId, p.Path, p.Name, c.installId)
	if err != nil {
		return nil, err
	}

	return map[string]any{"id": folderId}, nil
}

func (c *FilesCapability) getFileContent(params lazydata.LazyData) (any, error) {
	var p fileIdParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	content, err := c.fileOps.GetFileContent(c.installId, p.ID)
	if err != nil {
		return nil, err
	}

	return map[string]any{"content": string(content)}, nil
}

func (c *FilesCapability) getFileContentByPath(params lazydata.LazyData) (any, error) {
	var p filePathParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	content, err := c.fileOps.GetFileContentByPath(c.installId, p.Path, p.Name)
	if err != nil {
		return nil, err
	}

	return map[string]any{"content": string(content)}, nil
}

func (c *FilesCapability) getFileContentBase64(params lazydata.LazyData) (any, error) {
	var p fileIdParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	content, err := c.fileOps.GetFileContent(c.installId, p.ID)
	if err != nil {
		return nil, err
	}

	return map[string]any{"content": base64.StdEncoding.EncodeToString(content)}, nil
}

func (c *FilesCapability) getFileMeta(params lazydata.LazyData) (any, error) {
	var p fileIdParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	meta, err := c.fileOps.GetFileMeta(p.ID)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

func (c *FilesCapability) getFileMetaByPath(params lazydata.LazyData) (any, error) {
	var p filePathParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	meta, err := c.fileOps.GetFileMetaByPath(c.installId, p.Path, p.Name)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

func (c *FilesCapability) listFiles(params lazydata.LazyData) (any, error) {
	var p listFilesParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	files, err := c.fileOps.ListFiles(c.installId, p.Path)
	if err != nil {
		return nil, err
	}

	if files == nil {
		return []dbmodels.FileMeta{}, nil
	}

	return files, nil
}

func (c *FilesCapability) removeFile(params lazydata.LazyData) (any, error) {
	var p fileIdParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	err := c.fileOps.RemoveFile(c.installId, p.ID)
	if err != nil {
		return nil, err
	}

	return map[string]bool{"success": true}, nil
}

func (c *FilesCapability) updateFile(params lazydata.LazyData) (any, error) {
	var p updateFileParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	err := c.fileOps.UpdateFile(c.installId, p.ID, strings.NewReader(p.Content))
	if err != nil {
		return nil, err
	}

	return map[string]bool{"success": true}, nil
}

func (c *FilesCapability) updateFileMeta(params lazydata.LazyData) (any, error) {
	var p updateFileMetaParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	err := c.fileOps.UpdateFileMeta(c.installId, p.ID, p.Data)
	if err != nil {
		return nil, err
	}

	return map[string]bool{"success": true}, nil
}

// HTTP streaming handlers

func (c *FilesCapability) handleStreamFile(ctx *gin.Context, id int64) {
	err := c.fileOps.StreamFile(c.installId, id, ctx.Writer)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
	}
}

func (c *FilesCapability) handleStreamFileByPath(ctx *gin.Context, path string, name string) {
	err := c.fileOps.StreamFileByPath(c.installId, path, name, ctx.Writer)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
	}
}
