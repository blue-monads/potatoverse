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
	"github.com/blue-monads/turnix/backend/xtypes/models"
	"github.com/k0kubun/pp"
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

	packageVersions, err := e.db.GetPackageInstallOps().ListPackageVersionByIds(pversionIds)
	if err != nil {
		return err
	}

	pversionMap := make(map[int64]*dbmodels.PackageVersion)
	for _, pversion := range packageVersions {
		pversionMap[pversion.ID] = &pversion
	}

	for _, space := range spaces {

		packageVersion := pversionMap[space.InstalledId]
		if packageVersion == nil {
			e.logger.Warn("package version not found, skipping space", "space_id", space.ID, "installed_id", space.InstalledId)
			continue
		}

		routeOptions := models.PotatoRouteOptions{}
		err = json.Unmarshal([]byte(space.RouteOptions), &routeOptions)
		if err != nil {
			routeOptions.ServeFolder = "public"
			routeOptions.TrimPathPrefix = ""
			routeOptions.ForceHtmlExtension = false
			routeOptions.ForceIndexHtmlFile = true
			routeOptions.RouterType = "simple"

			e.logger.Warn("error unmarshalling route options, using default values", "error", err, "space_id", space.ID, "route_options", space.RouteOptions)

		}

		indexItem := &SpaceRouteIndexItem{
			installedId:      space.InstalledId,
			spaceId:          space.ID,
			routeOption:      routeOptions,
			packageVersionId: packageVersion.ID,
		}

		nextRoutingIndex[fmt.Sprintf("%d|_|%s", space.ID, space.NamespaceKey)] = indexItem

		exist := nextRoutingIndex[fmt.Sprintf("%s", space.NamespaceKey)]
		if exist == nil {
			nextRoutingIndex[space.NamespaceKey] = indexItem
		}

		if indexItem.routeOption.RouterType == "" {
			indexItem.routeOption.RouterType = "simple"
			indexItem.routeOption.ForceHtmlExtension = true
			indexItem.routeOption.ForceIndexHtmlFile = true
			indexItem.routeOption.ServeFolder = "public"

		}

		if routeOptions.RouterType == "dynamic" {
			indexItem.compiledTemplates = make(map[string]*template.Template)

			tempFolder, err := e.copyFolderToTemp(space.InstalledId, routeOptions.TemplateFolder)
			if err != nil {
				return err
			}

			// defer os.RemoveAll(tempFolder)

			for _, route := range routeOptions.Routes {
				if route.Type == "template" && route.File != "" {

					tmpl, err := template.ParseFiles(tempFolder + "/" + route.File)
					if err != nil {
						pp.Println("@err/5", err)
						return err
					}

					indexItem.compiledTemplates[route.File] = tmpl
				}
			}
		}

	}

	e.riLock.Lock()
	e.RoutingIndex = nextRoutingIndex
	e.riLock.Unlock()

	return nil
}

func (e *Engine) getIndex(spaceKey string, spaceId int64) *SpaceRouteIndexItem {
	e.riLock.RLock()
	defer e.riLock.RUnlock()

	if spaceId != 0 {
		key := fmt.Sprintf("%d|_|%s", spaceId, spaceKey)
		pp.Println("@getIndex/1", key)

		return e.RoutingIndex[key]
	}

	return e.RoutingIndex[spaceKey]
}

func (e *Engine) copyFolderToTemp(installedId int64, folder string) (string, error) {

	tempFolder := path.Join(os.TempDir(), "turnix", "packages", strconv.FormatInt(installedId, 10))

	os.MkdirAll(tempFolder, 0755)

	folderName := ""
	pathName := ""
	if strings.Contains(folder, "/") {
		folderName = filepath.Base(folder)
		pathName = filepath.Dir(folder)
	} else {
		folderName = folder
		pathName = ""
	}

	pp.Println("@folderName", folderName)
	pp.Println("@pathName", pathName)

	// folderFile, err := e.db.GetPackageFileMetaByPath(packageId, pathName, folderName)
	// if err != nil {
	// 	pp.Println("@err/1", err)
	// 	return "", err
	// }

	// err = copyFolder(e.db, tempFolder, folderFile)
	// if err != nil {
	// 	pp.Println("@err/2", err)
	// 	return "", err
	// }

	return path.Join(tempFolder, folderName), nil
}

func copyFolder(db datahub.Database, basePath string, folder *dbmodels.FileMeta) error {

	pp.Println("@copyFolder/1", basePath, folder.Path)

	// files, err := db.ListPackageFilesByPath(folder.OwnerID, folder.Name)
	// if err != nil {
	// 	pp.Println("@err/3", err)
	// 	return err
	// }

	// pp.Println("@copyFolder/2", files)

	// for _, file := range files {
	// 	pp.Println("@file", file.Name)
	// 	if file.IsFolder {
	// 		continue
	// 	}

	// 	pp.Println("@file/2", file.Path, file.Name)

	// 	filePath := path.Join(basePath, file.Path, file.Name)
	// 	//basePath + "/" + file.Path + "/" + file.Name

	// 	prePath := path.Join(basePath, file.Path)
	// 	err = os.MkdirAll(prePath, 0755)
	// 	if err != nil {
	// 		pp.Println("@err/4", err)
	// 		return err
	// 	}

	// 	pp.Println("@prePath", prePath)

	// 	tfile, err := os.Create(filePath)
	// 	if err != nil {
	// 		pp.Println("@err/5", err)
	// 		pp.Println("@err/5", err.Error())
	// 		return err
	// 	}
	// 	pp.Println("@file/3", filePath)

	// 	defer tfile.Close()

	// 	pp.Println("@file/4", file.PackageID, file.ID)

	// 	err = db.GetPackageFileStreaming(file.PackageID, file.ID, tfile)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	pp.Println("@file/5", err)

	// }

	// pp.Println("@copyFolder/3")

	return nil

}
