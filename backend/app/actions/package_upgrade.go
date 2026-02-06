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

// UpgradePackageResult is returned from upgrade operations, similar to InstallPackageResult for new installs.
type UpgradePackageResult struct {
	PackageVersionId int64  `json:"package_version_id"`
	UpdatePage       string `json:"update_page"`
	KeySpace         string `json:"key_space"`
	RootSpaceId      int64  `json:"root_space_id"`
}

func (c *Controller) UpgradePackageRepo(userId int64, repoSlug, version string, installedId int64) (*UpgradePackageResult, error) {

	pkg, err := c.database.GetPackageInstallOps().GetPackage(installedId)
	if err != nil {
		return nil, err
	}

	packageSlug := pkg.Slug

	rhub := c.engine.GetRepoHub()
	file, err := rhub.ZipPackage(repoSlug, packageSlug, version)
	if err != nil {
		return nil, err
	}

	defer os.Remove(file)

	return c.UpgradePackage(userId, file, installedId, true)

}

func (c *Controller) UpgradePackage(userId int64, file string, installedId int64, recreateArtifacts bool) (*UpgradePackageResult, error) {

	pvid, err := c.database.GetPackageInstallOps().UpdatePackage(installedId, file)
	if err != nil {
		return nil, err
	}

	rawPkg, err := xutils.GetPackageManifest(file)
	if err != nil {
		return nil, err
	}

	pkg := &models.PotatoPackage{}
	err = json.Unmarshal(rawPkg, pkg)
	if err != nil {
		return nil, err
	}

	oldSpaces, err := c.database.GetSpaceOps().ListSpacesByPackageId(installedId)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		for i, oldSpace := range oldSpaces {
			if oldSpace.NamespaceKey == space.Namespace {
				currentArtifactIndex = i
				break
			}
		}

		if space.Namespace == "" {
			return nil, errors.New("space namespace is required")
		}

		if currentArtifactIndex == -1 {
			spaceId, err := installArtifactSpace(c.database, userId, installedId, &space)
			if err != nil {
				return nil, err
			}

			c.logger.Info("space installed", "space_id", spaceId)
		} else {

			oldSpace := oldSpaces[currentArtifactIndex]

			if recreateArtifacts {

				routeOptions, err := json.Marshal(space.RouteOptions)
				if err != nil {
					return nil, err
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
					return nil, err
				}

			}

		}

	}

	pops := c.database.GetPackageInstallOps()
	err = pops.UpdateActiveInstallId(installedId, pvid)
	if err != nil {
		return nil, err
	}

	// delete old versions, keeping 3 latest versions

	allPVersions, err := pops.ListPackageVersionsByPackageId(installedId)
	if err != nil {
		return nil, err
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

	rootSpaceId := int64(0)
	for _, s := range oldSpaces {
		if s.NamespaceKey == pkg.Slug {
			rootSpaceId = s.ID
			break
		}
	}

	return &UpgradePackageResult{
		PackageVersionId: pvid,
		UpdatePage:       pkg.UpdatePage,
		KeySpace:         pkg.Slug,
		RootSpaceId:      rootSpaceId,
	}, nil

}
