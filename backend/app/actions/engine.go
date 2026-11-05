package actions

import (
	"errors"
	"fmt"

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

var (
	ErrUserNotAllowed = errors.New("you are not authorized to perform this action")
)

func (c *Controller) IsUserPackageAdmin(userId, installId int64) error {
	uops := c.database.GetUserOps()
	pkgOps := c.database.GetPackageInstallOps()
	sops := c.database.GetSpaceOps()

	user, err := uops.GetUser(userId)
	if err != nil {
		return err
	}

	pkg, err := pkgOps.GetPackage(installId)

	if user.Ugroup != "admin" && pkg.InstalledBy != userId {
		hasAdminScope := false

		users, err := sops.QuerySpaceUsers(pkg.ID, map[any]any{
			"user_id": userId,
		})
		if err == nil {
			for _, currUser := range users {
				if currUser.Scope == "core.admin" || currUser.Scope == "*" {
					hasAdminScope = true
					break
				}
			}
		}

		if !hasAdminScope && !c.userHasGroupPackageAdmin(user.Ugroup, pkg.ID) {
			return ErrUserNotAllowed
		}
	}

	return nil

}

type SpaceAuth struct {
	SpaceId int64 `json:"space_id"`
}

func (c *Controller) AuthorizeSpace(userId int64, req SpaceAuth) (string, error) {
	uops := c.database.GetUserOps()
	sops := c.database.GetSpaceOps()

	user, err := uops.GetUser(userId)
	if err != nil {
		return "", err
	}

	space, err := sops.GetSpace(req.SpaceId)
	if err != nil {
		return "", err
	}

	if user.Ugroup != "admin" && space.OwnerID != userId {
		hasAccess := false

		users, err := sops.QuerySpaceUsers(space.InstalledId, map[any]any{
			"user_id": userId,
		})
		if err == nil {
			for _, u := range users {
				if u.SpaceID == 0 || u.SpaceID == space.ID {
					hasAccess = true
					break
				}
			}
		}

		if !hasAccess && !c.userHasGroupSpaceAccess(user.Ugroup, space.InstalledId, space.ID) {
			return "", ErrUserNotAllowed
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

const PackageDevTokenPrefix = "ppsec_"

func (c *Controller) GeneratePackageDevToken(userId int64, packageId int64, epthermal bool) (string, error) {
	// Verify the user owns the package

	pkgOps := c.database.GetPackageInstallOps()

	pkg, err := pkgOps.GetPackage(packageId)
	if err != nil {
		return "", err
	}

	if pkg.DevToken != "" && !epthermal {
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

	token = PackageDevTokenPrefix + token

	if !epthermal {
		// Store the dev token in the database
		err = pkgOps.UpdatePackageDevToken(packageId, token)
		if err != nil {
			return "", err
		}
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

func (c *Controller) userHasGroupPackageAdmin(ugroupName string, installId int64) bool {
	if ugroupName == "" {
		return false
	}
	ugroup, err := c.database.GetUserOps().GetUserGroup(ugroupName)
	if err != nil || ugroup == nil {
		return false
	}
	groups, err := c.database.GetSpaceOps().QuerySpaceUserGroups(installId, map[any]any{
		"group_id": ugroup.ID,
	})
	if err != nil {
		return false
	}
	for _, currGroup := range groups {
		if currGroup.Scope == "core.admin" || currGroup.Scope == "*" {
			return true
		}
	}
	return false
}

func (c *Controller) userHasGroupSpaceAccess(ugroupName string, installId, spaceId int64) bool {
	if ugroupName == "" {
		return false
	}
	ugroup, err := c.database.GetUserOps().GetUserGroup(ugroupName)
	if err != nil || ugroup == nil {
		return false
	}
	groups, err := c.database.GetSpaceOps().QuerySpaceUserGroups(installId, map[any]any{
		"group_id": ugroup.ID,
	})
	if err != nil {
		return false
	}
	for _, g := range groups {
		if g.SpaceID == 0 || g.SpaceID == spaceId {
			return true
		}
	}
	return false
}

func (c *Controller) GetSpaceToken(userId int64, spaceId int64, namespaceKey string) (*dbmodels.Space, string, error) {
	sops := c.database.GetSpaceOps()
	var targetSpace *dbmodels.Space

	if spaceId > 0 {
		sp, err := sops.GetSpace(spaceId)
		if err != nil {
			return nil, "", err
		}
		targetSpace = sp
	} else if namespaceKey != "" {
		spaces, err := sops.ListSpaces()
		if err != nil {
			return nil, "", err
		}
		for i := range spaces {
			if spaces[i].NamespaceKey == namespaceKey {
				targetSpace = &spaces[i]
				break
			}
		}
		if targetSpace == nil {
			return nil, "", fmt.Errorf("space with namespace %q not found", namespaceKey)
		}
	} else {
		spaces, err := sops.ListSpaces()
		if err != nil {
			return nil, "", err
		}
		for i := range spaces {
			if spaces[i].SpaceType != "AppPlugin" {
				targetSpace = &spaces[i]
				break
			}
		}
		if targetSpace == nil && len(spaces) > 0 {
			targetSpace = &spaces[0]
		}
		if targetSpace == nil {
			return nil, "", errors.New("no spaces found")
		}
	}

	token, err := c.AuthorizeSpace(userId, SpaceAuth{SpaceId: targetSpace.ID})
	if err != nil {
		return nil, "", err
	}

	return targetSpace, token, nil
}

