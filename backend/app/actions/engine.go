package actions

import (
	"errors"
	"io"
	"net/http"
	"os"

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

func (c *Controller) GetEngineDebugData() map[string]any {
	return c.engine.GetDebugData()
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

func (c *Controller) InstallPackageByUrl(userId int64, url string) (int64, error) {

	tmpFile, err := os.CreateTemp("", "turnix-package-*.zip")
	if err != nil {
		return 0, err
	}
	defer os.Remove(tmpFile.Name())

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return 0, err
	}

	file := tmpFile.Name()

	return c.InstallPackageByFile(userId, file)

}

func (c *Controller) InstallPackageEmbed(userId int64, name string) (int64, error) {
	file, err := engine.ZipEPackage(name)
	if err != nil {
		return 0, err
	}

	defer os.Remove(file)

	return c.InstallPackageByFile(userId, file)
}

func (c *Controller) InstallPackageByFile(userId int64, file string) (int64, error) {
	packageId, err := c.database.InstallPackage(userId, file)
	if err != nil {
		return 0, err
	}

	pkg, err := c.database.GetPackage(packageId)
	if err != nil {
		return 0, err
	}

	spaceId, err := c.database.AddSpace(&models.Space{
		PackageID:     packageId,
		NamespaceKey:  pkg.Slug,
		OwnsNamespace: true,
		ExecutorType:  "luaz",
		SubType:       "space",
		OwnerID:       pkg.InstalledBy,
		IsInitilized:  false,
		IsPublic:      true,
	})

	if err != nil {
		return 0, err
	}

	c.engine.LoadRoutingIndex()

	c.logger.Info("space installed", "space_id", spaceId)

	return packageId, nil
}
