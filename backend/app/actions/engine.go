package actions

import (
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

	spaces, err := c.database.ListOwnSpaces(0, "")
	if err != nil {
		return nil, err
	}

	tpSpaces, err := c.database.ListThirdPartySpaces(userId, "")
	if err != nil {
		return nil, err
	}

	spaces = append(spaces, tpSpaces...)

	packageIds := make([]int64, 0, len(spaces))
	for _, space := range spaces {
		packageIds = append(packageIds, space.PackageID)
	}

	packages, err := c.database.ListPackagesByIds(packageIds)
	if err != nil {
		return nil, err
	}

	return &InstalledSpace{
		Spaces:   spaces,
		Packages: packages,
	}, nil

}
