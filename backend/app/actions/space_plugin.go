package actions

import (
	"errors"
	"fmt"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
)

type EnrichedSpacePlugin struct {
	dbmodels.SpacePlugin
	TargetNamespaceKey   string `json:"target_namespace_key,omitempty"`
	TargetPackageName    string `json:"target_package_name,omitempty"`
	TargetPackageVersion string `json:"target_package_version,omitempty"`
	TargetLoaderScript   string `json:"target_loader_script,omitempty"`
}

type AvailablePluginInfo struct {
	SpaceID          int64  `json:"space_id"`
	InstallID        int64  `json:"install_id"`
	NamespaceKey     string `json:"namespace_key"`
	LoaderScript     string `json:"loader_script"`
	PackageName      string `json:"package_name"`
	PackageInfo      string `json:"package_info"`
	PackageVersion   string `json:"package_version"`
	PackageAuthor    string `json:"package_author"`
	IsAlreadyPlugged bool   `json:"is_already_plugged"`
	CurrentPluginID  int64  `json:"current_plugin_id,omitempty"`
}

func (c *Controller) ListSpacePlugins(sourceInstallId, sourceSpaceId int64) ([]EnrichedSpacePlugin, error) {
	plugins, err := c.database.GetSpaceOps().ListSpacePlugins(sourceInstallId, sourceSpaceId)
	if err != nil {
		return nil, err
	}

	result := make([]EnrichedSpacePlugin, 0, len(plugins))
	for _, p := range plugins {
		item := EnrichedSpacePlugin{
			SpacePlugin: p,
		}

		targetSpace, err := c.database.GetSpaceOps().GetSpace(p.TargetSpaceID)
		if err == nil && targetSpace != nil {
			item.TargetNamespaceKey = targetSpace.NamespaceKey
			item.TargetLoaderScript = targetSpace.LoaderScript

			pkg, err := c.database.GetPackageInstallOps().GetPackage(targetSpace.InstalledId)
			if err == nil && pkg != nil {
				item.TargetPackageName = pkg.Name
				if pkg.ActiveInstallID != 0 {
					pver, err := c.database.GetPackageInstallOps().GetPackageVersion(pkg.ActiveInstallID)
					if err == nil && pver != nil {
						item.TargetPackageVersion = pver.Version
					}
				}
			}
		}

		result = append(result, item)
	}

	return result, nil
}

func (c *Controller) GetSpacePluginByID(sourceInstallId, pluginId int64) (*EnrichedSpacePlugin, error) {
	plugin, err := c.database.GetSpaceOps().GetSpacePlugin(pluginId)
	if err != nil {
		return nil, err
	}
	if plugin.SourceInstallID != sourceInstallId {
		return nil, errors.New("plugin not found for this package")
	}

	item := &EnrichedSpacePlugin{
		SpacePlugin: *plugin,
	}

	targetSpace, err := c.database.GetSpaceOps().GetSpace(plugin.TargetSpaceID)
	if err == nil && targetSpace != nil {
		item.TargetNamespaceKey = targetSpace.NamespaceKey
		item.TargetLoaderScript = targetSpace.LoaderScript

		pkg, err := c.database.GetPackageInstallOps().GetPackage(targetSpace.InstalledId)
		if err == nil && pkg != nil {
			item.TargetPackageName = pkg.Name
			if pkg.ActiveInstallID != 0 {
				pver, err := c.database.GetPackageInstallOps().GetPackageVersion(pkg.ActiveInstallID)
				if err == nil && pver != nil {
					item.TargetPackageVersion = pver.Version
				}
			}
		}
	}

	return item, nil
}

func (c *Controller) CreateSpacePlugin(sourceInstallId, sourceSpaceId int64, data map[string]any) (*dbmodels.SpacePlugin, error) {
	var targetInstallId int64
	var targetSpaceId int64

	if v, ok := data["target_install_id"].(float64); ok {
		targetInstallId = int64(v)
	} else if v, ok := data["target_install_id"].(int64); ok {
		targetInstallId = v
	}

	if v, ok := data["target_space_id"].(float64); ok {
		targetSpaceId = int64(v)
	} else if v, ok := data["target_space_id"].(int64); ok {
		targetSpaceId = v
	}

	if targetSpaceId == 0 {
		return nil, errors.New("target_space_id is required")
	}

	// Verify target space exists and is an AppPlugin
	targetSpace, err := c.database.GetSpaceOps().GetSpace(targetSpaceId)
	if err != nil {
		return nil, fmt.Errorf("target space not found: %w", err)
	}
	if targetSpace.SpaceType != "AppPlugin" {
		return nil, errors.New("target space must be of type AppPlugin")
	}

	if targetInstallId == 0 {
		targetInstallId = targetSpace.InstalledId
	}

	if sourceSpaceId == 0 {
		if v, ok := data["source_space_id"].(float64); ok {
			sourceSpaceId = int64(v)
		} else if v, ok := data["source_space_id"].(int64); ok {
			sourceSpaceId = v
		}
	}

	if sourceSpaceId == 0 {
		return nil, errors.New("source_space_id is required")
	}

	if sourceSpaceId == targetSpaceId {
		return nil, errors.New("cannot plug a space into itself")
	}

	extrameta, _ := data["extrameta"].(string)
	if extrameta == "" {
		extrameta = "{}"
	}

	pluginRecord := &dbmodels.SpacePlugin{
		SourceInstallID: sourceInstallId,
		SourceSpaceID:   sourceSpaceId,
		TargetInstallID: targetInstallId,
		TargetSpaceID:   targetSpaceId,
		ExtraMeta:       extrameta,
	}

	id, err := c.database.GetSpaceOps().AddSpacePlugin(pluginRecord)
	if err != nil {
		return nil, err
	}

	if c.engine != nil {
		c.engine.LoadRoutingIndexForPackages(sourceInstallId)
	}

	return c.database.GetSpaceOps().GetSpacePlugin(id)
}

func (c *Controller) UpdateSpacePluginByID(sourceInstallId, pluginId int64, data map[string]any) (*dbmodels.SpacePlugin, error) {
	plugin, err := c.database.GetSpaceOps().GetSpacePlugin(pluginId)
	if err != nil {
		return nil, err
	}
	if plugin.SourceInstallID != sourceInstallId {
		return nil, errors.New("plugin not found for this package")
	}

	updateMap := make(map[string]any)
	if extrameta, ok := data["extrameta"].(string); ok {
		updateMap["extrameta"] = extrameta
	}

	if len(updateMap) > 0 {
		err = c.database.GetSpaceOps().UpdateSpacePlugin(pluginId, updateMap)
		if err != nil {
			return nil, err
		}
		if c.engine != nil {
			c.engine.LoadRoutingIndexForPackages(sourceInstallId)
		}
	}

	return c.database.GetSpaceOps().GetSpacePlugin(pluginId)
}

func (c *Controller) DeleteSpacePluginByID(sourceInstallId, pluginId int64) error {
	plugin, err := c.database.GetSpaceOps().GetSpacePlugin(pluginId)
	if err != nil {
		return err
	}
	if plugin.SourceInstallID != sourceInstallId {
		return errors.New("plugin not found for this package")
	}

	err = c.database.GetSpaceOps().RemoveSpacePlugin(pluginId)
	if err != nil {
		return err
	}

	if c.engine != nil {
		c.engine.LoadRoutingIndexForPackages(sourceInstallId)
	}

	return nil
}

func (c *Controller) ListAvailableAppPlugins(sourceInstallId, sourceSpaceId int64) ([]AvailablePluginInfo, error) {
	appPlugins, err := c.database.GetSpaceOps().ListSpacesBySpaceType("AppPlugin")
	if err != nil {
		return nil, err
	}

	existingPlugins := make(map[int64]int64) // target_space_id -> plugin id
	if sourceSpaceId != 0 {
		plugins, err := c.database.GetSpaceOps().ListSpacePlugins(sourceInstallId, sourceSpaceId)
		if err == nil {
			for _, p := range plugins {
				existingPlugins[p.TargetSpaceID] = p.ID
			}
		}
	}

	result := make([]AvailablePluginInfo, 0, len(appPlugins))
	for _, sp := range appPlugins {
		if sp.ID == sourceSpaceId {
			continue
		}

		info := AvailablePluginInfo{
			SpaceID:      sp.ID,
			InstallID:    sp.InstalledId,
			NamespaceKey: sp.NamespaceKey,
			LoaderScript: sp.LoaderScript,
		}

		if pId, plugged := existingPlugins[sp.ID]; plugged {
			info.IsAlreadyPlugged = true
			info.CurrentPluginID = pId
		}

		pkg, err := c.database.GetPackageInstallOps().GetPackage(sp.InstalledId)
		if err == nil && pkg != nil {
			info.PackageName = pkg.Name
			if pkg.ActiveInstallID != 0 {
				pver, err := c.database.GetPackageInstallOps().GetPackageVersion(pkg.ActiveInstallID)
				if err == nil && pver != nil {
					info.PackageInfo = pver.Info
					info.PackageVersion = pver.Version
					info.PackageAuthor = pver.AuthorName
				}
			}
		}

		result = append(result, info)
	}

	return result, nil
}
