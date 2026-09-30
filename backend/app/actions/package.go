package actions

import (
	"encoding/json"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/repohub/repotypes"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

func (c *Controller) ListEPackages(repoSlug string) ([]repotypes.PotatoPackage, error) {
	repoHub := c.engine.GetRepoHub()

	return repoHub.ListPackages(repoSlug)
}

func (c *Controller) ListRepos() ([]xtypes.RepoOptions, error) {
	repoHub := c.engine.GetRepoHub()
	if repoHub == nil {
		return []xtypes.RepoOptions{}, nil
	}
	return repoHub.ListRepos(), nil
}

type InstalledSpace struct {
	Spaces   []dbmodels.Space          `json:"spaces"`
	Packages []dbmodels.PackageVersion `json:"packages"`
}

func (c *Controller) ListInstalledSpaces(userId int64) (*InstalledSpace, error) {
	sops := c.database.GetSpaceOps()
	uops := c.database.GetUserOps()

	ownspaces, err := sops.ListOwnSpaces(userId, "")
	if err != nil {
		return nil, err
	}

	tpSpaces, err := sops.ListThirdPartySpaces(userId, "")
	if err != nil {
		return nil, err
	}

	allCandidateSpaces := make([]dbmodels.Space, 0, len(ownspaces)+len(tpSpaces))
	allCandidateSpaces = append(allCandidateSpaces, ownspaces...)
	allCandidateSpaces = append(allCandidateSpaces, tpSpaces...)
	allCandidateSpaces = append(allCandidateSpaces, c.listPackageLevelUserSpaces(userId)...)

	// Check group access
	user, err := uops.GetUser(userId)
	if err == nil && user != nil {
		allCandidateSpaces = append(allCandidateSpaces, c.listGroupAccessibleSpaces(user.Ugroup)...)
	}

	installedIds := make([]int64, 0, len(allCandidateSpaces))
	for _, space := range allCandidateSpaces {
		installedIds = append(installedIds, space.InstalledId)
	}

	packages, err := c.database.GetPackageInstallOps().ListPackagesByIds(installedIds)
	if err != nil {
		return nil, err
	}

	packageVersions := make([]int64, 0, len(packages))

	for _, pkg := range packages {
		packageVersions = append(packageVersions, pkg.ActiveInstallID)
	}

	pversions, err := c.database.GetPackageInstallOps().ListPackageVersionByIds(packageVersions)
	if err != nil {
		return nil, err
	}

	finalSpaces := make([]dbmodels.Space, 0, len(allCandidateSpaces))
	hasPackageMap := make(map[int64]struct{})
	hasSpaceMap := make(map[int64]struct{})

	for _, pkg := range packages {
		hasPackageMap[pkg.ID] = struct{}{}
	}

	for _, space := range allCandidateSpaces {
		if space.SpaceType == "AppPlugin" {
			continue
		}
		if _, ok := hasPackageMap[space.InstalledId]; ok {
			if _, ok := hasSpaceMap[space.ID]; !ok {
				finalSpaces = append(finalSpaces, space)
				hasSpaceMap[space.ID] = struct{}{}
			}
		}
	}

	return &InstalledSpace{
		Spaces:   finalSpaces,
		Packages: pversions,
	}, nil

}

type InstalledPackageInfo struct {
	InstalledPackage *dbmodels.InstalledPackage `json:"installed_package"`
	Spaces           []dbmodels.Space           `json:"spaces"`
	PackageVersions  []dbmodels.PackageVersion  `json:"package_versions"`
}

func (c *Controller) GetInstalledPackageInfo(packageId int64) (*InstalledPackageInfo, error) {
	pkg, err := c.database.GetPackageInstallOps().GetPackage(packageId)
	if err != nil {
		return nil, err
	}

	// Get all versions for this package, not just the active one
	pversions, err := c.database.GetPackageInstallOps().ListPackageVersionsByPackageId(packageId)
	if err != nil {
		return nil, err
	}

	spaces, err := c.database.GetSpaceOps().ListSpacesByPackageId(packageId)
	if err != nil {
		return nil, err
	}

	return &InstalledPackageInfo{
		InstalledPackage: pkg,
		Spaces:           spaces,
		PackageVersions:  pversions,
	}, nil
}

// AvailableVersionsResponse is returned by ListPackageAvailableVersions.
type AvailableVersionsResponse struct {
	Versions       []string `json:"versions"`
	RepoSlug       string   `json:"repo_slug"`
	Name           string   `json:"name"`
	CurrentVersion string   `json:"current_version,omitempty"`
}

// ListPackageAvailableVersions returns versions available in the repo for the given installed package.
// The package must have been installed from a repo (install_repo set).
func (c *Controller) ListPackageAvailableVersions(packageId int64) (*AvailableVersionsResponse, error) {
	pkg, err := c.database.GetPackageInstallOps().GetPackage(packageId)
	if err != nil {
		return nil, err
	}
	if pkg.InstallRepo == "" {
		return &AvailableVersionsResponse{
			Versions:       nil,
			RepoSlug:       "",
			Name:           pkg.Name,
			CurrentVersion: "",
		}, nil
	}

	packages, err := c.engine.GetRepoHub().ListPackages(pkg.InstallRepo)
	if err != nil {
		return nil, err
	}

	var repoPkg *repotypes.PotatoPackage
	for i := range packages {
		if packages[i].Slug == pkg.Slug {
			repoPkg = &packages[i]
			break
		}
	}
	if repoPkg == nil || len(repoPkg.Versions) == 0 {
		return &AvailableVersionsResponse{
			Versions:       nil,
			RepoSlug:       pkg.InstallRepo,
			Name:           pkg.Name,
			CurrentVersion: "",
		}, nil
	}

	currentVersion := ""
	activeVer, err := c.database.GetPackageInstallOps().GetPackageVersion(pkg.ActiveInstallID)
	if err == nil && activeVer != nil {
		currentVersion = activeVer.Version
	}

	return &AvailableVersionsResponse{
		Versions:       repoPkg.Versions,
		RepoSlug:       pkg.InstallRepo,
		Name:           pkg.Name,
		CurrentVersion: currentVersion,
	}, nil
}

// GetEnvs returns package env vars as a flat JSON object (key -> value, single-level).
// Invalid or empty stored JSON returns an empty map.
func (c *Controller) GetEnvs(packageId int64) (map[string]string, error) {
	pkg, err := c.database.GetPackageInstallOps().GetPackage(packageId)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	if pkg.EnvVars == "" {
		return out, nil
	}
	if err := json.Unmarshal([]byte(pkg.EnvVars), &out); err != nil {
		return out, nil
	}
	return out, nil
}

// UpdateEnvs sets package env vars from a flat JSON object (key -> value, single-level only).
func (c *Controller) UpdateEnvs(packageId int64, envs map[string]string) error {
	raw, err := json.Marshal(envs)
	if err != nil {
		return err
	}
	return c.database.GetPackageInstallOps().UpdatePackageData(packageId, map[string]any{
		"env_vars": string(raw),
	})
}

func (c *Controller) listPackageLevelUserSpaces(userId int64) []dbmodels.Space {
	sops := c.database.GetSpaceOps()
	userDirectSpaceEntries, err := sops.QuerySpaceUsers(0, map[any]any{
		"user_id":  userId,
		"space_id": 0,
	})
	if err != nil {
		return nil
	}

	var spaces []dbmodels.Space
	for _, entry := range userDirectSpaceEntries {
		pkgSpaces, err := sops.ListSpacesByPackageId(entry.InstallID)
		if err == nil {
			spaces = append(spaces, pkgSpaces...)
		}
	}
	return spaces
}

func (c *Controller) listGroupAccessibleSpaces(ugroupName string) []dbmodels.Space {
	if ugroupName == "" {
		return nil
	}
	uops := c.database.GetUserOps()
	sops := c.database.GetSpaceOps()

	ugroup, err := uops.GetUserGroup(ugroupName)
	if err != nil || ugroup == nil {
		return nil
	}

	groupEntries, err := sops.QuerySpaceUserGroups(0, map[any]any{
		"group_id": ugroup.ID,
	})
	if err != nil {
		return nil
	}

	var spaces []dbmodels.Space
	for _, g := range groupEntries {
		if g.SpaceID > 0 {
			sp, err := sops.GetSpace(g.SpaceID)
			if err == nil && sp != nil {
				spaces = append(spaces, *sp)
			}
		} else {
			// Package-level grant: include all spaces of this install
			pkgSpaces, err := sops.ListSpacesByPackageId(g.InstallID)
			if err == nil {
				spaces = append(spaces, pkgSpaces...)
			}
		}
	}
	return spaces
}

