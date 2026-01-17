package engine

import (
	"encoding/json"
	"fmt"
	"html/template"
	"slices"
	"strings"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes/models"
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

func (e *Engine) LoadRoutingIndex() {
	e.fullReload <- struct{}{}
}

func (e *Engine) loadRoutingIndex() error {

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

func (e *Engine) LoadRoutingIndexForPackages(installedId int64) {
	e.reloadPackageIds <- installedId
}

func (e *Engine) loadRoutingIndexForPackages(installedIds ...int64) error {

	qq.Println("@loadRoutingIndexForPackages/1", installedIds)

	nextPartialIndex := make(map[string]*SpaceRouteIndexItem)

	// Get all spaces for the given installedIds
	allSpaces := make([]dbmodels.Space, 0)
	for _, installedId := range installedIds {
		spaces, err := e.db.GetSpaceOps().ListSpacesByPackageId(installedId)
		if err != nil {
			e.logger.Warn("failed to list spaces for installed package", "installed_id", installedId, "error", err)
			continue
		}
		allSpaces = append(allSpaces, spaces...)
	}

	qq.Println("@loadRoutingIndexForPackages/2", len(allSpaces))

	if len(allSpaces) == 0 {
		// No spaces to update, but still need to remove old entries
		e.riLock.Lock()
		// Remove entries for spaces that no longer exist or were removed
		for key, item := range e.RoutingIndex {
			if slices.Contains(installedIds, item.installedId) {
				delete(e.RoutingIndex, key)
			}
		}
		e.riLock.Unlock()
		return nil
	}

	qq.Println("@loadRoutingIndexForPackages/3", len(installedIds))

	// Get installed packages to find their ActiveInstallIDs
	installs, err := e.db.GetPackageInstallOps().ListPackagesByIds(installedIds)
	if err != nil {
		return err
	}

	pversionIds := make([]int64, 0, len(installs))
	for _, install := range installs {
		pversionIds = append(pversionIds, install.ActiveInstallID)
	}

	qq.Println("@loadRoutingIndexForPackages/4", len(pversionIds))

	packageVersions, err := e.db.GetPackageInstallOps().ListPackageVersionByIds(pversionIds)
	if err != nil {
		return err
	}

	qq.Println("@loadRoutingIndexForPackages/5", len(packageVersions))

	pversionMap := make(map[int64]*dbmodels.PackageVersion)
	for _, pversion := range packageVersions {
		pversionMap[pversion.InstallId] = &pversion
	}

	qq.Println("@loadRoutingIndexForPackages/6", len(pversionMap))

	// Build index items for affected spaces
	affectedSpaceIds := make(map[int64]struct{})
	for _, space := range allSpaces {
		affectedSpaceIds[space.ID] = struct{}{}

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

		key := fmt.Sprintf("%d|_|%s", space.ID, space.NamespaceKey)
		qq.Println("@loadRoutingIndexForPackages/7.1", key)

		nextPartialIndex[key] = indexItem

		qq.Println("@loadRoutingIndexForPackages/7.2", key)

		exist := nextPartialIndex[space.NamespaceKey]
		if exist == nil {
			nextPartialIndex[space.NamespaceKey] = indexItem
		}
	}

	qq.Println("@loadRoutingIndexForPackages/7", len(nextPartialIndex))

	e.riLock.Lock()
	// Remove old entries for affected spaces
	// First, collect keys to remove to avoid modifying map during iteration
	keysToRemove := make([]string, 0)
	for key, item := range e.RoutingIndex {
		// Remove if it belongs to an affected space
		if _, isAffected := affectedSpaceIds[item.spaceId]; isAffected {
			keysToRemove = append(keysToRemove, key)
		}
	}
	// Remove the collected keys
	for _, key := range keysToRemove {
		delete(e.RoutingIndex, key)
	}
	// Add new entries
	for key, item := range nextPartialIndex {
		// Space-specific keys are always updated
		if strings.HasPrefix(key, fmt.Sprintf("%d|_|", item.spaceId)) {
			e.RoutingIndex[key] = item
		} else {
			// For namespace keys, only add if they don't already exist
			// (preserving namespace keys from other packages)
			if e.RoutingIndex[key] == nil {
				e.RoutingIndex[key] = item
			}
		}
	}

	qq.Println("@loadRoutingIndexForPackages/8", len(e.RoutingIndex))

	e.riLock.Unlock()

	spaceIds := make([]int64, 0, len(affectedSpaceIds))
	for spaceId := range affectedSpaceIds {
		spaceIds = append(spaceIds, spaceId)
	}

	e.runtime.ClearExecs(spaceIds...)

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

		e.logger.Warn("failed to unmarshal route options", "space_id", space.ID, "installed_id", space.InstalledId, "error", err)

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

		e.logger.Warn("failed to set default route options", "space_id", space.ID, "installed_id", space.InstalledId)

	}

	if routeOptions.RouterType == "dynamic" {
		indexItem.compiledTemplates = make(map[string]*template.Template)

		for _, route := range routeOptions.Routes {
			if route.Type == "template" && route.File != "" {

				fileOps := e.db.GetPackageFileOps()

				asFS := fileOps.NewAsFS(packageVersion.ID, routeOptions.TemplateFolder)

				tmpl, err := template.ParseFS(asFS, route.File)
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

func (e *Engine) getIndexRetry(spaceKey string, spaceId int64) *SpaceRouteIndexItem {
	for i := 0; i < 5; i++ {
		index := e.getIndex(spaceKey, spaceId)
		if index != nil {
			return index
		}
		time.Sleep(2 * time.Second)
	}
	return nil

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
