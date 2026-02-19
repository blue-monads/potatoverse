package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	xutils "github.com/blue-monads/potatoverse/backend/utils"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes/models"
)

func (c *Controller) InstallPackageByUrl(userId int64, url string) (*InstallPackageResult, error) {

	tmpFile, err := os.CreateTemp("", "potato-package-*.zip")
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

	return c.InstallPackageByFile(userId, "", file)

}

func (c *Controller) InstallPackageRepo(userId int64, name string, repoSlug string) (*InstallPackageResult, error) {
	var file string
	var err error

	repoHub := c.engine.GetRepoHub()
	file, err = repoHub.ZipPackage(repoSlug, name, "")

	if err != nil {
		return nil, err
	}

	defer os.Remove(file)

	return c.InstallPackageByFile(userId, repoSlug, file)
}

func (c *Controller) InstallPackageByFile(userId int64, repo, file string) (*InstallPackageResult, error) {
	id, err := installPackageByFile(c.database, c.logger, userId, repo, file)
	if err != nil {
		return nil, err
	}

	c.engine.LoadRoutingIndexForPackages(id.InstalledId)

	return id, nil

}

type InstallPackageResult struct {
	InstalledId  int64             `json:"installed_id"`
	RootSpaceId  int64             `json:"root_space_id"`
	KeySpace     string            `json:"key_space"`
	SpecialPages map[string]string `json:"special_pages"`
}

func installPackageByFile(database datahub.Database, logger *slog.Logger, userId int64, repo, file string) (*InstallPackageResult, error) {

	pkgops := database.GetPackageInstallOps()

	installedId, err := pkgops.InstallPackage(userId, repo, file)
	if err != nil {
		return nil, err
	}

	rawPkg, err := xutils.GetPackageManifest(file)
	if err != nil {
		return nil, fmt.Errorf("failed to get package manifest: %w", err)
	}

	pkg := &models.PotatoPackage{}
	err = json.Unmarshal(rawPkg, pkg)
	if err != nil {
		return nil, err
	}

	err = checkSlug(pkg.Slug)
	if err != nil {
		return nil, err
	}

	rootSpaceId := int64(0)
	keySpace := pkg.Slug

	spaceMap := make(map[string]int64)
	foundRootSpace := false

	for _, space := range pkg.Spaces {
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

	for _, capability := range pkg.Capabilities {
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
	}

	ipkg, err := pkgops.GetPackage(installedId)
	if err != nil {
		return nil, err
	}

	vpkg, err := pkgops.GetPackageVersion(ipkg.ActiveInstallID)
	if err != nil {
		return nil, err
	}

	specialPages := map[string]string{}
	err = json.Unmarshal([]byte(vpkg.SpecialPages), &specialPages)
	if err != nil {
		return nil, err
	}

	return &InstallPackageResult{
		InstalledId:  installedId,
		RootSpaceId:  rootSpaceId,
		KeySpace:     keySpace,
		SpecialPages: specialPages,
	}, nil
}

func installCapability(database datahub.Database, installedId, spaceId int64, capability models.PotatoCapability) error {

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

func installArtifactSpace(database datahub.Database, userId, installedId int64, artifact *models.PotatoSpace) (int64, error) {
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

// private

// valid namespace should only contain letters, numbers, underscores and hyphens
var validNamespaceRegex = regexp.MustCompile(`^[a-zA-Z0-9_:-]+$`)
var validPkgSlugRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func checkSlug(slug string) error {
	if !validPkgSlugRegex.MatchString(slug) {
		return errors.New("package slug is invalid, it can only contain letters, numbers, and hyphens")
	}

	if strings.HasSuffix(slug, "-") {
		return errors.New("package slug must not end with a hyphen")
	}

	if strings.HasPrefix(slug, "-") {
		return errors.New("package slug must not start with a hyphen")
	}

	if strings.HasSuffix(slug, "_") {
		return errors.New("package slug must not end with an underscore")
	}

	if strings.HasPrefix(slug, "_") {
		return errors.New("package slug must not start with an underscore")
	}

	return nil
}
