package engine

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/xtypes/models"
	"github.com/k0kubun/pp"
)

type SpaceRouteIndexItem struct {
	packageId         int64
	spaceId           int64
	overlayForSpaceId int64
	routeOption       models.PotatoRouteOptions

	compiledTemplates map[string]*template.Template
}

type PluginRouteIndexItem struct {
	pluginId    int64
	packageId   int64
	spaceId     int64
	routeOption models.PotatoRouteOptions
}

func (e *Engine) LoadRoutingIndex() error {

	nextRoutingIndex := make(map[string]*SpaceRouteIndexItem)

	spaces, err := e.db.ListSpaces()
	if err != nil {
		return err
	}

	for _, space := range spaces {

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
			packageId:   space.PackageID,
			spaceId:     space.ID,
			routeOption: routeOptions,
		}

		if space.OwnsNamespace {
			nextRoutingIndex[space.NamespaceKey] = indexItem
		} else {
			nextRoutingIndex[fmt.Sprintf("%d|_|%s", space.ID, space.NamespaceKey)] = indexItem
		}

		if routeOptions.RouterType == "dynamic" {
			indexItem.compiledTemplates = make(map[string]*template.Template)

			tempFolder, err := e.copyFolderToTemp(space.PackageID, routeOptions.TemplateFolder)
			if err != nil {
				return err
			}

			defer os.RemoveAll(tempFolder)

			for _, route := range routeOptions.Routes {
				if route.Type == "template" && route.File != "" {

					tmpl, err := template.ParseFiles(tempFolder + "/" + route.File)
					if err != nil {
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
		spaceKey = fmt.Sprintf("%d|_|%s", spaceId, spaceKey)
	}

	return e.RoutingIndex[spaceKey]
}

func (e *Engine) copyFolderToTemp(packageId int64, folder string) (string, error) {
	tempFolder := os.TempDir() +
		"/turnix/packages/" +
		strconv.FormatInt(packageId, 10) +
		"/" + folder

	os.MkdirAll(tempFolder, 0755)

	folderName := filepath.Base(folder)
	pathName := filepath.Dir(folder)

	pp.Println("@folderName", folderName)
	pp.Println("@pathName", pathName)

	folderFile, err := e.db.GetPackageFileMetaByPath(packageId, pathName, folderName)
	if err != nil {
		return "", err
	}

	err = copyFolder(e.db, tempFolder, *folderFile)
	if err != nil {
		return "", err
	}

	return tempFolder, nil
}

func copyFolder(db datahub.Database, basePath string, folder dbmodels.PackageFile) error {

	files, err := db.ListPackageFilesByPath(folder.PackageID, folder.Path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsFolder {
			continue
		}

		filePath := basePath + "/" + file.Path + "/" + file.Name
		tfile, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer tfile.Close()

		err = db.GetPackageFileStreaming(file.PackageID, file.ID, tfile)
		if err != nil {
			return err
		}

	}

	return nil

}
