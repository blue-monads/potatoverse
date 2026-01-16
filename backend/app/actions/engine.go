package actions

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub"
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	xutils "github.com/blue-monads/potatoverse/backend/utils"
	"github.com/blue-monads/potatoverse/backend/utils/kosher"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes/models"
	"github.com/bwmarrin/snowflake"
	"github.com/tidwall/gjson"
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

func (c *Controller) InstallPackageByUrl(userId int64, url string) (*InstallPackageResult, error) {

	tmpFile, err := os.CreateTemp("", "turnix-package-*.zip")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return nil, err
	}

	file := tmpFile.Name()

	return c.InstallPackageByFile(userId, file)

}

func (c *Controller) GetPackage(packageId int64) (*dbmodels.InstalledPackage, error) {
	return c.database.GetPackageInstallOps().GetPackage(packageId)
}

func (c *Controller) GetPackageVersion(packageVersionId int64) (*dbmodels.PackageVersion, error) {
	return c.database.GetPackageInstallOps().GetPackageVersion(packageVersionId)
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

func (c *Controller) InstallPackageEmbed(userId int64, name string, repoSlug string) (*InstallPackageResult, error) {
	var file string
	var err error

	repoHub := c.engine.GetRepoHub()
	if repoHub != nil && repoSlug != "" {
		// Use RepoHub to get package from specific repo
		file, err = repohub.ZipEPackageFromRepo(repoHub, repoSlug, name, "")
	} else {
		// Fallback to default behavior for backward compatibility
		file, err = repohub.ZipEPackage(name)
	}

	if err != nil {
		return nil, err
	}

	defer os.Remove(file)

	return c.InstallPackageByFile(userId, file)
}

func (c *Controller) InstallPackageByFile(userId int64, file string) (*InstallPackageResult, error) {
	id, err := InstallPackageByFile(c.database, c.logger, userId, file)
	if err != nil {
		return nil, err
	}

	c.engine.LoadRoutingIndexForPackages(id.InstalledId)

	return id, nil

}

type InstallPackageResult struct {
	InstalledId int64  `json:"installed_id"`
	RootSpaceId int64  `json:"root_space_id"`
	KeySpace    string `json:"key_space"`
	InitPage    string `json:"init_page"`
}

// valid namespace should only contain letters, numbers, underscores and hyphens
var validNamespaceRegex = regexp.MustCompile(`^[a-zA-Z0-9_:-]+$`)
var validPkgSlugRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func InstallPackageByFile(database datahub.Database, logger *slog.Logger, userId int64, file string) (*InstallPackageResult, error) {

	installedId, err := database.GetPackageInstallOps().InstallPackage(userId, "embed", file)
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

	if !validPkgSlugRegex.MatchString(pkg.Slug) {
		return nil, errors.New("package slug is invalid, it can only contain letters, numbers, and hyphens")
	}

	if strings.HasSuffix(pkg.Slug, "-") {
		return nil, errors.New("package slug must not end with a hyphen")
	}

	if strings.HasPrefix(pkg.Slug, "-") {
		return nil, errors.New("package slug must not start with a hyphen")
	}

	if strings.HasSuffix(pkg.Slug, "_") {
		return nil, errors.New("package slug must not end with an underscore")
	}

	if strings.HasPrefix(pkg.Slug, "_") {
		return nil, errors.New("package slug must not start with an underscore")
	}

	rootSpaceId := int64(0)
	keySpace := pkg.Slug

	artifacts := gjson.GetBytes(rawPkg, "artifacts").Array()

	spaceMap := make(map[string]int64)

	foundRootSpace := false

	for index, artifact := range artifacts {
		kind := &pkg.Artifacts[index]

		if kind.Kind != "space" {
			continue
		}

		space := models.ArtifactSpace{}
		err = json.Unmarshal(kosher.Byte(artifact.Raw), &space)
		if err != nil {
			return nil, err
		}

		if space.Namespace == "" {
			return nil, errors.New("space namespace is required")
		}

		if space.Namespace == pkg.Slug {
			if foundRootSpace {
				return nil, errors.New("multiple root spaces found")
			}

			foundRootSpace = true
		} else {
			if !strings.HasPrefix(space.Namespace, pkg.Slug) {
				return nil, errors.New("space namespace must start with package slug (i.e. 'my-package:my-space')")
			} else if !strings.HasPrefix(space.Namespace, pkg.Slug+":") {
				return nil, errors.New("space namespace must start with package slug (i.e. 'my-package:my-space')")
			}
		}

		if !validNamespaceRegex.MatchString(space.Namespace) {
			return nil, errors.New("space namespace is invalid, it can only contain letters, numbers, underscores and hyphens")
		}

		if strings.HasSuffix(space.Namespace, ":") {
			return nil, errors.New("space namespace must not end with a colon")
		}

		if strings.HasPrefix(space.Namespace, ":") {
			return nil, errors.New("space namespace must not start with a colon")
		}

		spaceId, err := installArtifactSpace(database, userId, installedId, &space)
		if err != nil {
			return nil, err
		}

		spaceMap[space.Namespace] = spaceId

		if pkg.Slug == space.Namespace {
			rootSpaceId = spaceId
		}

		logger.Info("space installed", "space_id", spaceId)

	}

	qq.Println("@InstallPackageByFile/1", spaceMap)

	for index, artifact := range artifacts {
		kind := &pkg.Artifacts[index]

		switch kind.Kind {
		case "capability":

			capability := models.ArtifactCapability{}
			err = json.Unmarshal(kosher.Byte(artifact.Raw), &capability)
			if err != nil {
				return nil, err
			}

			qq.Println("@InstallPackageByFile/2", capability.Spaces)

			if len(capability.Spaces) != 0 {
				for _, space := range capability.Spaces {
					spaceId, ok := spaceMap[space]
					if !ok {
						return nil, errors.New("space not found")
					}

					err = installCapability(database, installedId, spaceId, capability)
					if err != nil {
						return nil, err
					}
				}
			} else {
				err = installCapability(database, installedId, 0, capability)
				if err != nil {
					return nil, err
				}
			}

		default:
			logger.Info("artifact is not a space", "artifact", artifact)
			continue
		}

	}

	return &InstallPackageResult{
		InstalledId: installedId,
		RootSpaceId: rootSpaceId,
		KeySpace:    keySpace,
		InitPage:    pkg.InitPage,
	}, nil
}

func installCapability(database datahub.Database, installedId, spaceId int64, capability models.ArtifactCapability) error {

	spaceOps := database.GetSpaceOps()

	options, err := json.Marshal(capability.Options)
	if err != nil {
		return err
	}

	return spaceOps.AddSpaceCapability(installedId, &dbmodels.SpaceCapability{
		InstallID:      installedId,
		Name:           capability.Name,
		CapabilityType: capability.Type,
		Options:        string(options),
		SpaceID:        spaceId,
		ExtraMeta:      "{}",
	})
}

func installArtifactSpace(database datahub.Database, userId, installedId int64, artifact *models.ArtifactSpace) (int64, error) {
	routeOptions, err := json.Marshal(artifact.RouteOptions)
	if err != nil {
		return 0, err
	}

	return database.GetSpaceOps().AddSpace(&dbmodels.Space{
		InstalledId:     installedId,
		NamespaceKey:    artifact.Namespace,
		ExecutorType:    artifact.ExecutorType,
		ExecutorSubType: artifact.ExecutorSubType,
		SpaceType:       "App",
		RouteOptions:    string(routeOptions),
		DevServePort:    int64(artifact.DevServePort),
		OwnerID:         userId,
		IsInitilized:    false,
		IsPublic:        true,
	})
}

func (c *Controller) UpgradePackage(userId int64, file string, installedId int64, recreateArtifacts bool) (int64, error) {

	pvid, err := c.database.GetPackageInstallOps().UpdatePackage(installedId, file)
	if err != nil {
		return 0, err
	}

	rawPkg, err := xutils.GetPackageManifest(file)
	if err != nil {
		return 0, err
	}

	pkg := &models.PotatoPackage{}
	err = json.Unmarshal(rawPkg, pkg)
	if err != nil {
		return 0, err
	}

	oldSpaces, err := c.database.GetSpaceOps().ListSpacesByPackageId(installedId)
	if err != nil {
		return 0, err
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
			return 0, err
		}

		for i, oldSpace := range oldSpaces {
			if oldSpace.NamespaceKey == space.Namespace {
				currentArtifactIndex = i
				break
			}
		}

		if space.Namespace == "" {
			return 0, errors.New("space namespace is required")
		}

		if currentArtifactIndex == -1 {
			spaceId, err := installArtifactSpace(c.database, userId, installedId, &space)
			if err != nil {
				return 0, err
			}

			c.logger.Info("space installed", "space_id", spaceId)
		} else {

			oldSpace := oldSpaces[currentArtifactIndex]

			if recreateArtifacts {

				routeOptions, err := json.Marshal(space.RouteOptions)
				if err != nil {
					return 0, err
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
					return 0, err
				}

			}

		}

	}

	pops := c.database.GetPackageInstallOps()
	if err != nil {
		return 0, err
	}
	err = pops.UpdateActiveInstallId(installedId, pvid)
	if err != nil {
		return 0, err
	}

	// delete old versions, keeping 3 latest versions

	allPVersions, err := pops.ListPackageVersionsByPackageId(installedId)
	if err != nil {
		return 0, err
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

	return pvid, nil

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
