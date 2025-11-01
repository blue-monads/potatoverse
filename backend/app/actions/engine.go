package actions

import (
	"archive/zip"
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

	packages, err := c.database.GetPackageInstallOps().ListPackageVersionByIds(installedIds)
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
		Packages: packages,
	}, nil

}

func (c *Controller) GetEngineDebugData() map[string]any {
	return c.engine.GetDebugData()
}

func (c *Controller) DeletePackage(userId int64, packageId int64) error {
	pkg, err := c.database.GetPackageInstallOps().GetPackage(packageId)
	if err != nil {
		return err
	}

	if pkg.InstalledBy != userId {
		return errors.New("you are not the owner of this package")
	}

	return c.database.GetPackageInstallOps().DeletePackage(packageId)

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

func (c *Controller) GetPackage(packageId int64) (*dbmodels.InstalledPackage, error) {
	return c.database.GetPackageInstallOps().GetPackage(packageId)
}

func (c *Controller) GeneratePackageDevToken(userId int64, packageId int64) (string, error) {
	// Verify the user owns the package
	pkg, err := c.database.GetPackageInstallOps().GetPackage(packageId)
	if err != nil {
		return "", err
	}

	if pkg.InstalledBy != userId {
		return "", errors.New("you are not the owner of this package")
	}

	// Generate the dev token
	return c.signer.SignPackageDev(&signer.PackageDevClaim{
		InstallPackageId: packageId,
		UserId:           userId,
		Typeid:           signer.ToekenPackageDev,
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

func InstallPackageByFile(database datahub.Database, logger *slog.Logger, userId int64, file string) (int64, error) {

	installedId, err := database.GetPackageInstallOps().InstallPackage(userId, "embed", file)
	if err != nil {
		return 0, err
	}

	pkg, err := readPackagePotatoManifestFromZip(file)
	if err != nil {
		return 0, err
	}

	for _, artifact := range pkg.Artifacts {
		if artifact.Kind != "space" {
			logger.Info("artifact is not a space", "artifact", artifact)
			continue
		}

		spaceId, err := installArtifact(database, userId, installedId, artifact)
		if err != nil {
			return 0, err
		}

		logger.Info("space installed", "space_id", spaceId)

	}

	return installedId, nil
}

func installArtifact(database datahub.Database, userId, installedId int64, artifact models.PotatoArtifact) (int64, error) {
	routeOptions, err := json.Marshal(artifact.RouteOptions)
	if err != nil {
		return 0, err
	}

	mcpOptions, err := json.Marshal(artifact.McpOptions)
	if err != nil {
		return 0, err
	}

	return database.GetSpaceOps().AddSpace(&dbmodels.Space{
		InstalledId:       installedId,
		NamespaceKey:      artifact.Namespace,
		ExecutorType:      artifact.ExecutorType,
		ExecutorSubType:   artifact.ExecutorSubType,
		SpaceType:         "App",
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

func readPackagePotatoManifestFromZip(zipFile string) (*models.PotatoPackage, error) {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		if file.Name == "potato.json" {
			jsonFile, err := file.Open()
			if err != nil {
				return nil, err
			}

			pkg := &models.PotatoPackage{}
			json.NewDecoder(jsonFile).Decode(&pkg)
			if err != nil {
				return nil, err
			}

			return pkg, nil
		}
	}

	return nil, errors.New("potato.json not found")
}

func (c *Controller) UpgradePackage(userId int64, file string, installedId int64, recreateArtifacts bool) (int64, error) {

	pvid, err := c.database.GetPackageInstallOps().UpdatePackage(installedId, file)
	if err != nil {
		return 0, err
	}

	pkg, err := readPackagePotatoManifestFromZip(file)
	if err != nil {
		return 0, err
	}

	oldSpaces, err := c.database.GetSpaceOps().ListSpacesByPackageId(installedId)
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
			spaceId, err := installArtifact(c.database, userId, installedId, artifact)
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

				c.database.GetSpaceOps().UpdateSpace(oldSpace.ID, map[string]any{
					"namespace_key":       artifact.Namespace,
					"executor_type":       artifact.ExecutorType,
					"executor_sub_type":   artifact.ExecutorSubType,
					"space_type":          "App",
					"route_options":       string(routeOptions),
					"mcp_enabled":         artifact.McpOptions.Enabled,
					"mcp_definition_file": artifact.McpOptions.DefinitionFile,
					"mcp_options":         string(mcpOptions),
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

	c.engine.LoadRoutingIndex()

	return pvid, nil

}
