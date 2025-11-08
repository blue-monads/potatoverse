package engine

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes/models"
)

type SpaceRouteIndexItem struct {
	installedId       int64
	packageVersionId  int64
	spaceId           int64
	overlayForSpaceId int64
	routeOption       models.PotatoRouteOptions

	compiledTemplates map[string]*template.Template
}

type PluginRouteIndexItem struct {
	pluginId         int64
	installedId      int64
	packageVersionId int64
	spaceId          int64
	routeOption      models.PotatoRouteOptions
}

func (e *Engine) LoadRoutingIndex() error {

	nextRoutingIndex := make(map[string]*SpaceRouteIndexItem)

	spaces, err := e.db.GetSpaceOps().ListSpaces()
	if err != nil {
		return err
	}

	installs, err := e.db.GetPackageInstallOps().ListPackages()
	if err != nil {
		return err
	}

	pversionIds := make([]int64, 0, len(installs))
	for _, install := range installs {
		pversionIds = append(pversionIds, install.ActiveInstallID)
	}

	qq.Println("@pversionIds", pversionIds)

	packageVersions, err := e.db.GetPackageInstallOps().ListPackageVersionByIds(pversionIds)
	if err != nil {
		return err
	}

	qq.Println("@packageVersions", len(packageVersions))

	pversionMap := make(map[int64]*dbmodels.PackageVersion)
	for _, pversion := range packageVersions {
		pversionMap[pversion.InstallId] = &pversion
	}

	for _, space := range spaces {

		packageVersion := pversionMap[space.InstalledId]
		if packageVersion == nil {
			e.logger.Warn("package version not found, skipping space", "space_id", space.ID, "installed_id", space.InstalledId)
			continue
		}

		indexItem, err := e.buildIndexItem(&space, packageVersion)
		if err != nil {
			e.logger.Warn("failed to build index item", "space_id", space.ID, "installed_id", space.InstalledId, "error", err)
			continue
		}

		nextRoutingIndex[fmt.Sprintf("%d|_|%s", space.ID, space.NamespaceKey)] = indexItem

		exist := nextRoutingIndex[fmt.Sprintf("%s", space.NamespaceKey)]
		if exist == nil {
			nextRoutingIndex[space.NamespaceKey] = indexItem
		}

	}

	e.riLock.Lock()
	e.RoutingIndex = nextRoutingIndex
	e.riLock.Unlock()

	return nil
}

func (e *Engine) buildIndexItem(space *dbmodels.Space, packageVersion *dbmodels.PackageVersion) (*SpaceRouteIndexItem, error) {

	routeOptions := models.PotatoRouteOptions{}
	err := json.Unmarshal([]byte(space.RouteOptions), &routeOptions)
	if err != nil {
		routeOptions.ServeFolder = "public"
		routeOptions.TrimPathPrefix = ""
		routeOptions.ForceHtmlExtension = false
		routeOptions.ForceIndexHtmlFile = true
		routeOptions.RouterType = "simple"

	}

	indexItem := &SpaceRouteIndexItem{
		installedId:      space.InstalledId,
		spaceId:          space.ID,
		routeOption:      routeOptions,
		packageVersionId: packageVersion.ID,
	}

	if indexItem.routeOption.RouterType == "" {
		indexItem.routeOption.RouterType = "simple"
		indexItem.routeOption.ForceHtmlExtension = true
		indexItem.routeOption.ForceIndexHtmlFile = true
		indexItem.routeOption.ServeFolder = "public"

	}

	if routeOptions.RouterType == "dynamic" {
		indexItem.compiledTemplates = make(map[string]*template.Template)

		tempFolder, err := e.copyFolderToTemp(packageVersion.ID, routeOptions.TemplateFolder)
		if err != nil {
			return nil, err
		}

		// defer os.RemoveAll(tempFolder)

		for _, route := range routeOptions.Routes {
			if route.Type == "template" && route.File != "" {

				tmpl, err := template.ParseFiles(tempFolder + "/" + route.File)
				if err != nil {
					qq.Println("@err/5", err)
					return nil, err
				}

				indexItem.compiledTemplates[route.File] = tmpl
			}
		}
	}

	return indexItem, nil
}

func (e *Engine) getIndex(spaceKey string, spaceId int64) *SpaceRouteIndexItem {
	e.riLock.RLock()
	defer e.riLock.RUnlock()

	if spaceId != 0 {
		key := fmt.Sprintf("%d|_|%s", spaceId, spaceKey)
		qq.Println("@getIndex/1", key)

		return e.RoutingIndex[key]
	}

	return e.RoutingIndex[spaceKey]
}

func (e *Engine) copyFolderToTemp(packageVersionId int64, folder string) (string, error) {

	tempFolder := path.Join(os.TempDir(), "turnix", "packages", strconv.FormatInt(packageVersionId, 10))

	os.MkdirAll(tempFolder, 0755)

	// Normalize the folder path
	folder = strings.TrimSuffix(folder, "/")
	folder = strings.TrimPrefix(folder, "/")

	qq.Println("@folder", folder)

	fileOps := e.db.GetPackageFileOps()

	// List all files in this folder path
	// Note: folders aren't explicitly stored, so we list files by path
	files, err := fileOps.ListFiles(packageVersionId, folder)
	if err != nil {
		qq.Println("@err/1", err)
		return "", fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		e.logger.Warn("no files found in template folder", "package_version_id", packageVersionId, "folder", folder)
	}

	// Create target directory
	folderName := filepath.Base(folder)
	if folderName == "." || folderName == "" {
		folderName = folder
		if folderName == "" {
			folderName = "root"
		}
	}
	targetPath := path.Join(tempFolder, folderName)
	os.MkdirAll(targetPath, 0755)

	// Copy all files (handle recursive copying for subdirectories)
	err = e.copyFilesRecursive(fileOps, packageVersionId, files, folder, targetPath)
	if err != nil {
		qq.Println("@err/2", err)
		return "", fmt.Errorf("failed to copy files: %w", err)
	}

	return targetPath, nil
}

func (e *Engine) copyFilesRecursive(fileOps datahub.FileOps, packageVersionId int64, files []dbmodels.FileMeta, sourceFolderPath string, targetBasePath string) error {

	qq.Println("@copyFilesRecursive/1", targetBasePath, "sourceFolder:", sourceFolderPath, "files:", len(files))

	for _, file := range files {
		qq.Println("@file", file.Name, "path:", file.Path, "is_folder:", file.IsFolder)

		// Skip folder entries (if any exist)
		if file.IsFolder {
			// Recurse into subfolder by listing files in it
			subfolderPath := path.Join(file.Path, file.Name)
			subfiles, err := fileOps.ListFiles(packageVersionId, subfolderPath)
			if err != nil {
				e.logger.Warn("failed to list subfolder files", "error", err, "subfolder_path", subfolderPath)
				continue
			}
			targetSubfolderPath := path.Join(targetBasePath, file.Name)
			err = e.copyFilesRecursive(fileOps, packageVersionId, subfiles, subfolderPath, targetSubfolderPath)
			if err != nil {
				return err
			}
			continue
		}

		// Calculate target file path
		// Files in the source folder will have Path == sourceFolderPath
		// We need to preserve any subdirectory structure
		var targetFilePath string
		if file.Path == sourceFolderPath {
			// File is directly in the source folder
			targetFilePath = path.Join(targetBasePath, file.Name)
		} else if strings.HasPrefix(file.Path, sourceFolderPath+"/") {
			// File is in a subdirectory
			// Get the relative path from sourceFolderPath
			relPath := strings.TrimPrefix(file.Path, sourceFolderPath+"/")
			targetFilePath = path.Join(targetBasePath, relPath, file.Name)
		} else {
			// This shouldn't happen, but handle it
			targetFilePath = path.Join(targetBasePath, file.Name)
		}

		// Create parent directories
		targetDir := filepath.Dir(targetFilePath)
		err := os.MkdirAll(targetDir, 0755)
		if err != nil {
			qq.Println("@err/4", err)
			return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}

		// Create the file
		tfile, err := os.Create(targetFilePath)
		if err != nil {
			qq.Println("@err/5", err)
			return fmt.Errorf("failed to create file %s: %w", targetFilePath, err)
		}

		qq.Println("@copying file", targetFilePath, "file.ID:", file.ID, "file.Size:", file.Size, "file.Path:", file.Path, "file.Name:", file.Name, "file.StoreType:", file.StoreType)

		// Try getting file content as bytes first to verify it exists
		content, err := fileOps.GetFileContentByPath(packageVersionId, file.Path, file.Name)
		if err != nil {
			tfile.Close()
			os.Remove(targetFilePath) // Clean up empty file
			qq.Println("@err/getcontent", err)
			return fmt.Errorf("failed to get file content %s (path: %s, name: %s): %w", targetFilePath, file.Path, file.Name, err)
		}

		qq.Println("@got file content", "size:", len(content), "bytes")

		if len(content) == 0 && file.Size > 0 {
			tfile.Close()
			os.Remove(targetFilePath)
			qq.Println("@warn", "file content is empty but size is", file.Size)
			return fmt.Errorf("file content is empty but file size is %d bytes", file.Size)
		}

		// Write content to file
		bytesWritten, err := tfile.Write(content)
		if err != nil {
			tfile.Close()
			os.Remove(targetFilePath)
			qq.Println("@err/write", err)
			return fmt.Errorf("failed to write file %s: %w", targetFilePath, err)
		}

		if bytesWritten != len(content) {
			tfile.Close()
			os.Remove(targetFilePath)
			qq.Println("@err/write", "incomplete write")
			return fmt.Errorf("incomplete write: wrote %d of %d bytes", bytesWritten, len(content))
		}

		qq.Println("@wrote", bytesWritten, "bytes to file")

		// Ensure data is written to disk
		err = tfile.Sync()
		if err != nil {
			tfile.Close()
			qq.Println("@err/sync", err)
			return fmt.Errorf("failed to sync file %s: %w", targetFilePath, err)
		}

		err = tfile.Close()
		if err != nil {
			qq.Println("@err/7", err)
			return fmt.Errorf("failed to close file %s: %w", targetFilePath, err)
		}

		// Verify file was written correctly
		info, err := os.Stat(targetFilePath)
		if err != nil {
			qq.Println("@err/stat", err)
			return fmt.Errorf("failed to stat copied file %s: %w", targetFilePath, err)
		}

		qq.Println("@file copied successfully", targetFilePath, "size:", info.Size(), "expected:", len(content))

		if info.Size() != int64(len(content)) {
			return fmt.Errorf("file size mismatch: expected %d bytes, got %d bytes", len(content), info.Size())
		}
	}

	qq.Println("@copyFilesRecursive/2", "done")

	return nil

}
