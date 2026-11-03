package rspace

import (
	"encoding/json"
	"fmt"
	"html/template"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes/models"
)

type SpaceRouteIndexItem struct {
	InstalledId      int64
	PackageVersionId int64
	SpaceId          int64
	RouteOption      models.PotatoRouteOptions

	AllPluginLoaderScript string // concat of all plugin loader scripts

	CompiledTemplates map[string]*template.Template
}

func (r *SpaceRouter) LoadRoutingIndex() {
	r.fullReload <- struct{}{}
}

func (r *SpaceRouter) loadRoutingIndex() error {

	nextRoutingIndex := make(map[string]*SpaceRouteIndexItem)

	spaces, err := r.db.GetSpaceOps().ListSpaces()
	if err != nil {
		return err
	}

	installs, err := r.db.GetPackageInstallOps().ListPackages()
	if err != nil {
		return err
	}

	pversionIds := make([]int64, 0, len(installs))
	for _, install := range installs {
		pversionIds = append(pversionIds, install.ActiveInstallID)
	}

	qq.Println("@pversionIds", pversionIds)

	packageVersions, err := r.db.GetPackageInstallOps().ListPackageVersionByIds(pversionIds)
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
			r.logger.Warn("package version not found, skipping space", "space_id", space.ID, "installed_id", space.InstalledId)
			continue
		}

		indexItem, err := r.BuildIndexItem(&space, packageVersion)
		if err != nil {
			r.logger.Warn("failed to build index item", "space_id", space.ID, "installed_id", space.InstalledId, "error", err)
			continue
		}

		nextRoutingIndex[fmt.Sprintf("%d|_|%s", space.ID, space.NamespaceKey)] = indexItem
		nextRoutingIndex[fmt.Sprintf("%d", space.ID)] = indexItem

		exist := nextRoutingIndex[fmt.Sprintf("%s", space.NamespaceKey)]
		if exist == nil {
			nextRoutingIndex[space.NamespaceKey] = indexItem
		}

	}

	r.riLock.Lock()
	r.RoutingIndex = nextRoutingIndex
	r.riLock.Unlock()

	return nil
}

func (r *SpaceRouter) LoadRoutingIndexForPackages(installedId int64) {
	r.reloadPackageIds <- installedId
}

func (r *SpaceRouter) loadRoutingIndexForPackages(installedIds ...int64) error {

	qq.Println("@loadRoutingIndexForPackages/1", installedIds)

	nextPartialIndex := make(map[string]*SpaceRouteIndexItem)

	// Get all spaces for the given installedIds
	allSpaces := make([]dbmodels.Space, 0)
	for _, installedId := range installedIds {
		spaces, err := r.db.GetSpaceOps().ListSpacesByPackageId(installedId)
		if err != nil {
			r.logger.Warn("failed to list spaces for installed package", "installed_id", installedId, "error", err)
			continue
		}
		allSpaces = append(allSpaces, spaces...)
	}

	qq.Println("@loadRoutingIndexForPackages/2", len(allSpaces))

	if len(allSpaces) == 0 {
		// No spaces to update, but still need to remove old entries
		r.riLock.Lock()
		// Remove entries for spaces that no longer exist or were removed
		for key, item := range r.RoutingIndex {
			if slices.Contains(installedIds, item.InstalledId) {
				delete(r.RoutingIndex, key)
			}
		}
		r.riLock.Unlock()
		return nil
	}

	qq.Println("@loadRoutingIndexForPackages/3", len(installedIds))

	// Get installed packages to find their ActiveInstallIDs
	installs, err := r.db.GetPackageInstallOps().ListPackagesByIds(installedIds)
	if err != nil {
		return err
	}

	pversionIds := make([]int64, 0, len(installs))
	for _, install := range installs {
		pversionIds = append(pversionIds, install.ActiveInstallID)
	}

	qq.Println("@loadRoutingIndexForPackages/4", len(pversionIds))

	packageVersions, err := r.db.GetPackageInstallOps().ListPackageVersionByIds(pversionIds)
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
			r.logger.Warn("package version not found, skipping space", "space_id", space.ID, "installed_id", space.InstalledId)
			continue
		}

		indexItem, err := r.BuildIndexItem(&space, packageVersion)
		if err != nil {
			r.logger.Warn("failed to build index item", "space_id", space.ID, "installed_id", space.InstalledId, "error", err)
			continue
		}

		key := fmt.Sprintf("%d|_|%s", space.ID, space.NamespaceKey)
		qq.Println("@loadRoutingIndexForPackages/7.1", key)

		nextPartialIndex[key] = indexItem
		nextPartialIndex[fmt.Sprintf("%d", space.ID)] = indexItem

		qq.Println("@loadRoutingIndexForPackages/7.2", key)

		exist := nextPartialIndex[space.NamespaceKey]
		if exist == nil {
			nextPartialIndex[space.NamespaceKey] = indexItem
		}
	}

	qq.Println("@loadRoutingIndexForPackages/7", len(nextPartialIndex))

	r.riLock.Lock()
	// Remove old entries for affected spaces
	// First, collect keys to remove to avoid modifying map during iteration
	keysToRemove := make([]string, 0)
	for key, item := range r.RoutingIndex {
		// Remove if it belongs to an affected space
		if _, isAffected := affectedSpaceIds[item.SpaceId]; isAffected {
			keysToRemove = append(keysToRemove, key)
		}
	}
	// Remove the collected keys
	for _, key := range keysToRemove {
		delete(r.RoutingIndex, key)
	}
	// Add new entries
	for key, item := range nextPartialIndex {
		// Space-specific keys are always updated
		if strings.HasPrefix(key, fmt.Sprintf("%d|_|", item.SpaceId)) || key == fmt.Sprintf("%d", item.SpaceId) {
			r.RoutingIndex[key] = item
		} else {
			// For namespace keys, only add if they don't already exist
			// (preserving namespace keys from other packages)
			if r.RoutingIndex[key] == nil {
				r.RoutingIndex[key] = item
			}
		}
	}

	qq.Println("@loadRoutingIndexForPackages/8", len(r.RoutingIndex))

	r.riLock.Unlock()

	spaceIds := make([]int64, 0, len(affectedSpaceIds))
	for spaceId := range affectedSpaceIds {
		spaceIds = append(spaceIds, spaceId)
	}

	if r.runtime != nil {
		r.runtime.ClearExecs(spaceIds...)
	}

	return nil
}

func (r *SpaceRouter) BuildIndexItem(space *dbmodels.Space, packageVersion *dbmodels.PackageVersion) (*SpaceRouteIndexItem, error) {

	routeOptions := models.PotatoRouteOptions{}
	err := json.Unmarshal([]byte(space.RouteOptions), &routeOptions)
	if err != nil {
		routeOptions.ServeFolder = "public"
		routeOptions.TrimPathPrefix = ""
		routeOptions.ForceHtmlExtension = false
		routeOptions.ForceIndexHtmlFile = true
		routeOptions.RouterType = "simple"

		r.logger.Warn("failed to unmarshal route options", "space_id", space.ID, "installed_id", space.InstalledId, "error", err)

	}

	indexItem := &SpaceRouteIndexItem{
		InstalledId:      space.InstalledId,
		SpaceId:          space.ID,
		RouteOption:      routeOptions,
		PackageVersionId: packageVersion.ID,
	}

	if indexItem.RouteOption.RouterType == "" {
		indexItem.RouteOption.RouterType = "simple"
		indexItem.RouteOption.ForceHtmlExtension = true
		indexItem.RouteOption.ForceIndexHtmlFile = true
		indexItem.RouteOption.ServeFolder = "public"

		r.logger.Warn("failed to set default route options", "space_id", space.ID, "installed_id", space.InstalledId)

	}

	if routeOptions.RouterType == "dynamic" {
		indexItem.CompiledTemplates = make(map[string]*template.Template)

		for _, route := range routeOptions.Routes {
			if route.Type == "template" && route.File != "" {

				fileOps := r.db.GetPackageFileOps()

				asFS := fileOps.NewAsFS(packageVersion.ID, routeOptions.TemplateFolder)

				tmpl, err := template.ParseFS(asFS, route.File)
				if err != nil {
					qq.Println("@err/5", err)
					return nil, err
				}

				indexItem.CompiledTemplates[route.File] = tmpl
			}
		}
	}

	// Load and concatenate all plugged plugin loader scripts
	var pluginScripts strings.Builder
	plugins, err := r.db.GetSpaceOps().ListSpacePlugins(space.InstalledId, space.ID)
	if err == nil && len(plugins) > 0 {

		const base = `(() => {

			let potatoRegistryFactory = {};
			const REG_KEY = "__pototo_registry_factory__";
			const SPACE_META_KEY = "__potato_space_meta__";

			if (window[REG_KEY]) {
				potatoRegistryFactory = window[REG_KEY];
			} else {
				window[REG_KEY] = potatoRegistryFactory;
				
				window[SPACE_META_KEY] = {
					space_id: %d,
					package_version_id: %d,
					installed_id: %d,
					namespace_key: "%s"
				};

			}

		})();`

		pluginScripts.WriteString(fmt.Sprintf(base, space.ID, packageVersion.ID, space.InstalledId, space.NamespaceKey))

		for _, plug := range plugins {
			targetSpace, err := r.db.GetSpaceOps().GetSpace(plug.TargetSpaceID)
			if err != nil || targetSpace == nil {
				continue
			}
			if targetSpace.LoaderScript == "" {
				continue
			}

			targetPkg, err := r.db.GetPackageInstallOps().GetPackage(targetSpace.InstalledId)
			if err != nil || targetPkg == nil || targetPkg.ActiveInstallID == 0 {
				continue
			}

			dir := path.Dir(targetSpace.LoaderScript)
			if dir == "." {
				dir = ""
			}
			name := path.Base(targetSpace.LoaderScript)

			scriptBytes, err := r.db.GetPackageFileOps().GetFileContentByPath(targetPkg.ActiveInstallID, dir, name)
			if err != nil {
				r.logger.Warn("failed to read plugin loader script", "target_space_id", targetSpace.ID, "loader_script", targetSpace.LoaderScript, "error", err)
				continue
			}

			if pluginScripts.Len() > 0 {
				pluginScripts.WriteString("\n\n")
			}
			pluginScripts.WriteString(fmt.Sprintf("// --- Plugin: %s (space: %d) ---\n", targetSpace.NamespaceKey, targetSpace.ID))
			pluginScripts.Write(scriptBytes)
		}
	}
	indexItem.AllPluginLoaderScript = pluginScripts.String()

	return indexItem, nil
}

func (r *SpaceRouter) buildIndexItem(space *dbmodels.Space, packageVersion *dbmodels.PackageVersion) (*SpaceRouteIndexItem, error) {
	return r.BuildIndexItem(space, packageVersion)
}

func (r *SpaceRouter) GetPluginLoaderScript(spaceKey string) string {
	if r == nil {
		return ""
	}
	index := r.getIndex(spaceKey, 0)
	if index != nil {
		return index.AllPluginLoaderScript
	}

	if id, err := strconv.ParseInt(spaceKey, 10, 64); err == nil && id > 0 {
		r.riLock.RLock()
		defer r.riLock.RUnlock()
		for _, item := range r.RoutingIndex {
			if item.InstalledId == id || item.SpaceId == id {
				return item.AllPluginLoaderScript
			}
		}
	}

	return ""
}

func (r *SpaceRouter) getIndexRetry(spaceKey string, spaceId int64) *SpaceRouteIndexItem {
	for i := 0; i < 5; i++ {
		index := r.getIndex(spaceKey, spaceId)
		if index != nil {
			return index
		}
		time.Sleep(2 * time.Second)
	}
	return nil

}

func (r *SpaceRouter) getIndex(spaceKey string, spaceId int64) *SpaceRouteIndexItem {
	r.riLock.RLock()
	defer r.riLock.RUnlock()

	if spaceId != 0 {
		if spaceKey != "" {
			key := fmt.Sprintf("%d|_|%s", spaceId, spaceKey)
			qq.Println("@getIndex/1", key)
			if item := r.RoutingIndex[key]; item != nil {
				return item
			}
		}

		key := fmt.Sprintf("%d", spaceId)
		if item := r.RoutingIndex[key]; item != nil {
			return item
		}

		for _, item := range r.RoutingIndex {
			if item.SpaceId == spaceId {
				return item
			}
		}

		return nil
	}

	return r.RoutingIndex[spaceKey]
}
