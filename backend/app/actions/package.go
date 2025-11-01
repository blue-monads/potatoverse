package actions

import (
	"github.com/blue-monads/turnix/backend/engine"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/xtypes/models"
)

func (c *Controller) ListEPackages() ([]models.PotatoPackage, error) {
	return engine.ListEPackages()
}

type InstalledSpace struct {
	Spaces   []dbmodels.Space          `json:"spaces"`
	Packages []dbmodels.PackageVersion `json:"packages"`
}

func (c *Controller) ListInstalledSpaces(userId int64) (*InstalledSpace, error) {

	ownspaces, err := c.database.GetSpaceOps().ListOwnSpaces(userId, "")
	if err != nil {
		return nil, err
	}

	tpSpaces, err := c.database.GetSpaceOps().ListThirdPartySpaces(userId, "")
	if err != nil {
		return nil, err
	}

	installedIds := make([]int64, 0, len(ownspaces)+len(tpSpaces))
	for _, space := range ownspaces {
		installedIds = append(installedIds, space.InstalledId)
	}

	for _, space := range tpSpaces {
		installedIds = append(installedIds, space.InstalledId)
	}

	packages, err := c.database.GetPackageInstallOps().ListPackagesByIds(installedIds)
	if err != nil {
		return nil, err
	}

	packageVersions := make([]int64, 0, len(packages))

	for _, pkg := range packages {
		packageVersions = append(packageVersions, pkg.ActiveInstallID)
	}

	pversions, err := c.database.GetPackageInstallOps().ListPackageVersionByIds(packageVersions)
	if err != nil {
		return nil, err
	}

	finalSpaces := make([]dbmodels.Space, 0, len(ownspaces)+len(tpSpaces))
	hasPackageMap := make(map[int64]struct{})

	for _, pkg := range packages {
		hasPackageMap[pkg.ID] = struct{}{}
	}

	for _, space := range ownspaces {
		if _, ok := hasPackageMap[space.InstalledId]; ok {
			finalSpaces = append(finalSpaces, space)
		}
	}

	for _, space := range tpSpaces {
		if _, ok := hasPackageMap[space.InstalledId]; ok {
			finalSpaces = append(finalSpaces, space)
		}
	}

	return &InstalledSpace{
		Spaces:   finalSpaces,
		Packages: pversions,
	}, nil

}

type InstalledPackageInfo struct {
	InstalledPackage *dbmodels.InstalledPackage `json:"installed_package"`
	Spaces           []dbmodels.Space           `json:"spaces"`
	PackageVersions  []dbmodels.PackageVersion  `json:"package_versions"`
}

func (c *Controller) GetInstalledPackageInfo(packageId int64) (*InstalledPackageInfo, error) {
	pkg, err := c.database.GetPackageInstallOps().GetPackage(packageId)
	if err != nil {
		return nil, err
	}

	// Get all versions for this package, not just the active one
	pversions, err := c.database.GetPackageInstallOps().ListPackageVersionsByPackageId(packageId)
	if err != nil {
		return nil, err
	}

	spaces, err := c.database.GetSpaceOps().ListSpacesByPackageId(packageId)
	if err != nil {
		return nil, err
	}

	return &InstalledPackageInfo{
		InstalledPackage: pkg,
		Spaces:           spaces,
		PackageVersions:  pversions,
	}, nil
}
