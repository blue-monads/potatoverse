package actions

import (
	"encoding/json"
	"errors"
	"os"
	"sort"

	xutils "github.com/blue-monads/potatoverse/backend/utils"
	"github.com/blue-monads/potatoverse/backend/utils/kosher"
	"github.com/blue-monads/potatoverse/backend/xtypes/models"
	"github.com/tidwall/gjson"
)

func (c *Controller) UpgradePackageRepo(userId int64, repoSlug, version string, installedId int64) (int64, error) {

	pkg, err := c.database.GetPackageInstallOps().GetPackage(installedId)
	if err != nil {
		return 0, err
	}

	packageSlug := pkg.Slug

	rhub := c.engine.GetRepoHub()
	file, err := rhub.ZipPackage(repoSlug, packageSlug, version)
	if err != nil {
		return 0, err
	}

	defer os.Remove(file)

	return c.UpgradePackage(userId, file, installedId, true)

}

func (c *Controller) UpgradePackage(userId int64, file string, installedId int64, recreateArtifacts bool) (int64, error) {

	pvid, err := c.database.GetPackageInstallOps().UpdatePackage(installedId, file)
	if err != nil {
		return 0, err
	}

	rawPkg, err := xutils.GetPackageManifest(file)
	if err != nil {
		return 0, err
	}

	pkg := &models.PotatoPackage{}
	err = json.Unmarshal(rawPkg, pkg)
	if err != nil {
		return 0, err
	}

	oldSpaces, err := c.database.GetSpaceOps().ListSpacesByPackageId(installedId)
	if err != nil {
		return 0, err
	}

	artifacts := gjson.GetBytes(rawPkg, "artifacts").Array()

	for index, artifact := range artifacts {
		kind := &pkg.Artifacts[index]

		if kind.Kind != "space" {
			continue
		}

		currentArtifactIndex := -1

		space := models.ArtifactSpace{}
		err = json.Unmarshal(kosher.Byte(artifact.Raw), &space)
		if err != nil {
			return 0, err
		}

		for i, oldSpace := range oldSpaces {
			if oldSpace.NamespaceKey == space.Namespace {
				currentArtifactIndex = i
				break
			}
		}

		if space.Namespace == "" {
			return 0, errors.New("space namespace is required")
		}

		if currentArtifactIndex == -1 {
			spaceId, err := installArtifactSpace(c.database, userId, installedId, &space)
			if err != nil {
				return 0, err
			}

			c.logger.Info("space installed", "space_id", spaceId)
		} else {

			oldSpace := oldSpaces[currentArtifactIndex]

			if recreateArtifacts {

				routeOptions, err := json.Marshal(space.RouteOptions)
				if err != nil {
					return 0, err
				}

				c.database.GetSpaceOps().UpdateSpace(oldSpace.ID, map[string]any{
					"namespace_key":     space.Namespace,
					"executor_type":     space.ExecutorType,
					"executor_sub_type": space.ExecutorSubType,
					"space_type":        "App",
					"route_options":     string(routeOptions),
				})

			} else {
				err = c.database.GetSpaceOps().UpdateSpace(oldSpace.ID, map[string]any{
					"install_id": installedId,
				})
				if err != nil {
					return 0, err
				}

			}

		}

	}

	pops := c.database.GetPackageInstallOps()
	if err != nil {
		return 0, err
	}
	err = pops.UpdateActiveInstallId(installedId, pvid)
	if err != nil {
		return 0, err
	}

	// delete old versions, keeping 3 latest versions

	allPVersions, err := pops.ListPackageVersionsByPackageId(installedId)
	if err != nil {
		return 0, err
	}

	if len(allPVersions) > 3 {
		sort.Slice(allPVersions, func(i, j int) bool {
			return allPVersions[i].ID > allPVersions[j].ID
		})

		for _, pversion := range allPVersions[3:] {
			err = pops.DeletePackageVersion(pversion.ID)
			if err != nil {
				c.logger.Error("failed to delete old package version", "error", err)
				continue
			}
		}
	}

	c.engine.LoadRoutingIndexForPackages(installedId)

	return pvid, nil

}
