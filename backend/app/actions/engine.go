package actions

import (
	"errors"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/bwmarrin/snowflake"
)

var (
	snode *snowflake.Node
)

func init() {
	_snode, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}

	snode = _snode
}

func (c *Controller) GetEngineDebugData() map[string]any {
	return c.engine.GetDebugData()
}

func (c *Controller) DeletePackage(userId int64, packageId int64) error {

	qq.Println("@DeletePackage/1", userId, packageId)

	pkg, err := c.database.GetPackageInstallOps().GetPackage(packageId)
	if err != nil {
		return err
	}

	qq.Println("@DeletePackage/2", pkg)

	if pkg.InstalledBy != userId {

		return errors.New("you are not the owner of this package")
	}

	qq.Println("@DeletePackage/3", "you are the owner of this package")

	err = c.database.GetPackageInstallOps().DeletePackage(packageId)
	if err != nil {
		return err
	}

	qq.Println("@DeletePackage/4", "deleting package")

	pkvVersions, err := c.database.GetPackageInstallOps().ListPackageVersionsByPackageId(packageId)
	if err != nil {
		return err
	}

	qq.Println("@DeletePackage/5", pkvVersions)

	spaceDb := c.database.GetSpaceOps()
	pkgInstallDb := c.database.GetPackageInstallOps()

	qq.Println("@DeletePackage/6")

	for _, pkvVersion := range pkvVersions {

		qq.Println("@DeletePackage/7", pkvVersion)

		err = pkgInstallDb.DeletePackageVersion(pkvVersion.ID)
		if err != nil {
			return err
		}

		qq.Println("@DeletePackage/8", "deleting package version")

		spaces, err := spaceDb.ListSpacesByPackageId(pkvVersion.InstallId)
		if err != nil {
			return err
		}

		qq.Println("@DeletePackage/9")

		for _, space := range spaces {
			err = spaceDb.RemoveSpace(space.ID)
			if err != nil {
				return err
			}
		}

	}

	return nil

}

type SpaceAuth struct {
	PackageId int64 `json:"package_id"`
	SpaceId   int64 `json:"space_id"`
}

func (c *Controller) AuthorizeSpace(userId int64, req SpaceAuth) (string, error) {

	space, err := c.database.GetSpaceOps().GetSpace(req.SpaceId)
	if err != nil {
		return "", err
	}

	if space.OwnerID != userId {
		_, err := c.database.GetSpaceOps().GetSpaceUserScope(userId, req.SpaceId)
		if err != nil {
			return "", errors.New("you are not authorized to access this space")
		}
	}

	return c.signer.SignSpace(&signer.SpaceClaim{
		SpaceId:   req.SpaceId,
		UserId:    userId,
		Typeid:    signer.TokenTypeSpace,
		InstallId: space.InstalledId,
		SessionId: snode.Generate().Int64(),
	})

}

func (c *Controller) GetPackage(packageId int64) (*dbmodels.InstalledPackage, error) {
	return c.database.GetPackageInstallOps().GetPackage(packageId)
}

func (c *Controller) GetPackageVersion(packageVersionId int64) (*dbmodels.PackageVersion, error) {
	return c.database.GetPackageInstallOps().GetPackageVersion(packageVersionId)
}

func (c *Controller) GeneratePackageDevToken(userId int64, packageId int64) (string, error) {
	// Verify the user owns the package

	pkgOps := c.database.GetPackageInstallOps()

	pkg, err := pkgOps.GetPackage(packageId)
	if err != nil {
		return "", err
	}

	if pkg.DevToken != "" {
		return pkg.DevToken, nil
	}

	if pkg.InstalledBy != userId {
		return "", errors.New("you are not the owner of this package")
	}

	// Generate the dev token
	token, err := c.signer.SignPackageDev(&signer.PackageDevClaim{
		InstallPackageId: packageId,
		UserId:           userId,
		Typeid:           signer.ToekenPackageDev,
	})
	if err != nil {
		return "", err
	}

	// Store the dev token in the database
	err = pkgOps.UpdatePackageDevToken(packageId, token)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (c *Controller) GetSpaceSpec(installedId int64) ([]byte, error) {
	spec, err := c.database.GetPackageInstallOps().GetPackage(installedId)
	if err != nil {
		return nil, err
	}

	activeInstallVersionId := spec.ActiveInstallID

	content, err := c.database.GetPackageFileOps().GetFileContentByPath(activeInstallVersionId, "", "spec.json")
	if err != nil {
		return nil, err
	}

	return content, nil
}
