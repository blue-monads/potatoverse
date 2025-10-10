package actions

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/blue-monads/turnix/backend/engine"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/xtypes/models"
	"github.com/rs/xid"
)

func (c *Controller) ListEPackages() ([]models.PotatoPackage, error) {
	return engine.ListEPackages()
}

type InstalledSpace struct {
	Spaces   []dbmodels.Space   `json:"spaces"`
	Packages []dbmodels.Package `json:"packages"`
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

	finalSpaces := make([]dbmodels.Space, 0, len(ownspaces)+len(tpSpaces))
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

type SpaceAuth struct {
	PackageId int64 `json:"package_id"`
	SpaceId   int64 `json:"space_id"`
}

func (c *Controller) AuthorizeSpace(userId int64, req SpaceAuth) (string, error) {

	space, err := c.database.GetSpace(req.SpaceId)
	if err != nil {
		return "", err
	}

	if space.OwnerID != userId {
		_, err := c.database.GetSpaceUserScope(userId, req.SpaceId)
		if err != nil {
			return "", errors.New("you are not authorized to access this space")
		}
	}

	return c.signer.SignSpace(&signer.SpaceClaim{
		SpaceId: req.SpaceId,
		UserId:  userId,
		XID:     xid.New().String(),
		Typeid:  signer.TokenTypeSpace,
	})

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

func (c *Controller) GetPackage(packageId int64) (*dbmodels.Package, error) {
	return c.database.GetPackage(packageId)
}

func (c *Controller) InstallPackageEmbed(userId int64, name string) (int64, error) {
	file, err := engine.ZipEPackage(name)
	if err != nil {
		return 0, err
	}

	// defer os.Remove(file)

	return c.InstallPackageByFile(userId, file)
}

func (c *Controller) InstallPackageByFile(userId int64, file string) (int64, error) {
	packageId, err := c.database.InstallPackage(userId, file)
	if err != nil {
		return 0, err
	}

	buf := bytes.Buffer{}
	err = c.database.GetPackageFileStreamingByPath(packageId, "", "potato.json", &buf)
	if err != nil {
		return 0, err
	}

	pkg := &models.PotatoPackage{}
	err = json.Unmarshal(buf.Bytes(), pkg)
	if err != nil {
		return 0, err
	}

	for _, artifact := range pkg.Artifacts {
		if artifact.Kind != "space" {
			c.logger.Info("artifact is not a space", "artifact", artifact)
			continue
		}

		routeOptions, err := json.Marshal(artifact.RouteOptions)
		if err != nil {
			return 0, err
		}

		mcpOptions, err := json.Marshal(artifact.McpOptions)
		if err != nil {
			return 0, err
		}

		spaceId, err := c.database.AddSpace(&dbmodels.Space{
			PackageID:         packageId,
			NamespaceKey:      artifact.Namespace,
			OwnsNamespace:     true,
			ExecutorType:      artifact.ExecutorType,
			SubType:           artifact.SubType,
			RouteOptions:      string(routeOptions),
			McpEnabled:        artifact.McpOptions.Enabled,
			McpDefinitionFile: artifact.McpOptions.DefinitionFile,
			McpOptions:        string(mcpOptions),
			DevServePort:      int64(artifact.DevServePort),
			OwnerID:           userId,
			IsInitilized:      false,
			IsPublic:          true,
		})

		if err != nil {
			return 0, err
		}

		c.logger.Info("space installed", "space_id", spaceId)

	}

	c.engine.LoadRoutingIndex()

	return packageId, nil
}
