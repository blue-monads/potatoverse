package actions

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/blue-monads/turnix/backend/engine"
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/xtypes/models"
	"github.com/jaevor/go-nanoid"
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

func (c *Controller) GeneratePackageDevToken(userId int64, packageId int64) (string, error) {
	// Verify the user owns the package
	pkg, err := c.database.GetPackage(packageId)
	if err != nil {
		return "", err
	}

	if pkg.InstalledBy != userId {
		return "", errors.New("you are not the owner of this package")
	}

	// Generate the dev token
	return c.signer.SignPackageDev(&signer.PackageDevClaim{
		PackageXID: pkg.XID,
		UserId:     userId,
		Typeid:     signer.ToekenPackageDev,
	})
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
	id, err := InstallPackageByFile(c.database, c.logger, userId, file)
	if err != nil {
		return 0, err
	}

	c.engine.LoadRoutingIndex()

	return id, nil

}

var packageXIDGenerator, _ = nanoid.ASCII(8)

func InstallPackageByFile(database datahub.Database, logger *slog.Logger, userId int64, file string) (int64, error) {
	pxid := packageXIDGenerator()

	packageId, err := database.InstallPackage(userId, file, pxid)
	if err != nil {
		return 0, err
	}

	pkg, err := readPackagePotatoManifest(database, packageId)
	if err != nil {
		return 0, err
	}

	for _, artifact := range pkg.Artifacts {
		if artifact.Kind != "space" {
			logger.Info("artifact is not a space", "artifact", artifact)
			continue
		}

		spaceId, err := installArtifact(database, userId, packageId, pxid, artifact)
		if err != nil {
			return 0, err
		}

		logger.Info("space installed", "space_id", spaceId)

	}

	return packageId, nil
}

func installArtifact(database datahub.Database, userId, packageId int64, pxid string, artifact models.PotatoArtifact) (int64, error) {
	routeOptions, err := json.Marshal(artifact.RouteOptions)
	if err != nil {
		return 0, err
	}

	mcpOptions, err := json.Marshal(artifact.McpOptions)
	if err != nil {
		return 0, err
	}

	return database.AddSpace(&dbmodels.Space{
		PackageID:         packageId,
		PackageXID:        pxid,
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
}

func readPackagePotatoManifest(database datahub.Database, packageId int64) (*models.PotatoPackage, error) {

	buf := bytes.Buffer{}
	err := database.GetPackageFileStreamingByPath(packageId, "", "potato.json", &buf)
	if err != nil {
		return nil, err
	}

	pkg := &models.PotatoPackage{}
	err = json.Unmarshal(buf.Bytes(), pkg)
	if err != nil {
		return nil, err
	}

	return pkg, nil
}

func (c *Controller) UpgradePackage(userId int64, file, pxid string, recreateArtifacts bool) (int64, error) {

	packages, err := c.database.ListPackagesByXID(pxid)
	if err != nil {
		return 0, err
	}

	oldPackageId := packages[0].ID

	for _, pkg := range packages {
		if oldPackageId > pkg.ID {
			oldPackageId = pkg.ID
		}
	}

	packageId, err := c.database.InstallPackage(userId, file, pxid)
	if err != nil {
		return 0, err
	}

	pkg, err := readPackagePotatoManifest(c.database, packageId)
	if err != nil {
		return 0, err
	}

	oldSpaces, err := c.database.ListSpacesByPackageId(oldPackageId)
	if err != nil {
		return 0, err
	}

	for _, artifact := range pkg.Artifacts {
		if artifact.Kind != "space" {
			continue
		}

		currentArtifactIndex := -1

		for i, oldSpace := range oldSpaces {
			if oldSpace.NamespaceKey == artifact.Namespace {
				currentArtifactIndex = i
				break
			}
		}

		if currentArtifactIndex == -1 {
			spaceId, err := installArtifact(c.database, userId, packageId, pxid, artifact)
			if err != nil {
				return 0, err
			}

			c.logger.Info("space installed", "space_id", spaceId)
		} else {

			oldSpace := oldSpaces[currentArtifactIndex]

			if recreateArtifacts {

				routeOptions, err := json.Marshal(artifact.RouteOptions)
				if err != nil {
					return 0, err
				}

				mcpOptions, err := json.Marshal(artifact.McpOptions)
				if err != nil {
					return 0, err
				}

				c.database.UpdateSpace(oldSpace.ID, map[string]any{
					"package_id":          packageId,
					"namespace_key":       artifact.Namespace,
					"executor_type":       artifact.ExecutorType,
					"sub_type":            artifact.SubType,
					"route_options":       string(routeOptions),
					"mcp_enabled":         artifact.McpOptions.Enabled,
					"mcp_definition_file": artifact.McpOptions.DefinitionFile,
					"mcp_options":         string(mcpOptions),
				})

			} else {
				err = c.database.UpdateSpace(oldSpace.ID, map[string]any{
					"package_id": packageId,
				})
				if err != nil {
					return 0, err
				}

			}

		}

	}

	c.engine.LoadRoutingIndex()

	return packageId, nil

}
