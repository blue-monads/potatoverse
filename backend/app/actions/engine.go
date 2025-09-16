package actions

import (
	"errors"

	"github.com/blue-monads/turnix/backend/engine"
	"github.com/blue-monads/turnix/backend/services/datahub/models"
)

func (c *Controller) ListEPackages() ([]engine.EPackage, error) {
	return engine.ListEPackages()
}

type InstalledSpace struct {
	Spaces   []models.Space   `json:"spaces"`
	Packages []models.Package `json:"packages"`
}

func (c *Controller) ListInstalledSpaces(userId int64) (*InstalledSpace, error) {

	ownspaces, err := c.database.ListOwnSpaces(userId, "")
	if err != nil {
		return nil, err
	}

	tpSpaces, err := c.database.ListThirdPartySpaces(userId, "")
	if err != nil {
		return nil, err
	}

	packageIds := make([]int64, 0, len(ownspaces)+len(tpSpaces))
	for _, space := range ownspaces {
		packageIds = append(packageIds, space.PackageID)
	}

	for _, space := range tpSpaces {
		packageIds = append(packageIds, space.PackageID)
	}

	packages, err := c.database.ListPackagesByIds(packageIds)
	if err != nil {
		return nil, err
	}

	finalSpaces := make([]models.Space, 0, len(ownspaces)+len(tpSpaces))
	hasPackageMap := make(map[int64]struct{})

	for _, pkg := range packages {
		hasPackageMap[pkg.ID] = struct{}{}
	}

	for _, space := range ownspaces {
		if _, ok := hasPackageMap[space.PackageID]; ok {
			finalSpaces = append(finalSpaces, space)
		}
	}

	for _, space := range tpSpaces {
		if _, ok := hasPackageMap[space.PackageID]; ok {
			finalSpaces = append(finalSpaces, space)
		}
	}

	return &InstalledSpace{
		Spaces:   finalSpaces,
		Packages: packages,
	}, nil

}

func (c *Controller) DeletePackage(userId int64, packageId int64) error {
	pkg, err := c.database.GetPackage(packageId)
	if err != nil {
		return err
	}

	if pkg.InstalledBy != userId {
		return errors.New("you are not the owner of this package")
	}

	return c.database.DeletePackage(packageId)

}
