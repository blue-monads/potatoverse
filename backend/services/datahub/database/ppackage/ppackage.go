package ppackage

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blue-monads/turnix/backend/services/datahub/database/file"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	xutils "github.com/blue-monads/turnix/backend/utils"
	"github.com/upper/db/v4"
)

type PackageInstallOperations struct {
	db      db.Session
	fileOps *file.FileOperations
}

func NewPackageInstallOperations(db db.Session, pfileOps *file.FileOperations) *PackageInstallOperations {
	return &PackageInstallOperations{
		db:      db,
		fileOps: pfileOps,
	}
}

func (d *PackageInstallOperations) InstallPackage(userId int64, repo, filePath string) (int64, error) {
	pkgManifest, err := xutils.ReadPackageManifestFromZip(filePath)
	if err != nil {
		return 0, err
	}

	t := time.Now()

	// Create InstalledPackage record
	installedPackage := &dbmodels.InstalledPackage{
		Name:            pkgManifest.Name,
		InstallRepo:     repo,
		UpdateUrl:       pkgManifest.UpdateUrl,
		ActiveInstallID: userId,
		InstalledBy:     userId,
		InstalledAt:     &t,
	}

	// Insert package
	result, err := d.installedPackagesTable().Insert(installedPackage)
	if err != nil {
		return 0, err
	}

	installId := result.ID().(int64)

	// Create a version entry
	version := &dbmodels.PackageVersion{
		InstallId:     installId,
		Name:          pkgManifest.Name,
		Slug:          pkgManifest.Slug,
		Info:          pkgManifest.Info,
		Tags:          strings.Join(pkgManifest.Tags, ","),
		Version:       pkgManifest.Version,
		FormatVersion: pkgManifest.FormatVersion,
		AuthorName:    pkgManifest.AuthorName,
		AuthorEmail:   pkgManifest.AuthorEmail,
		AuthorSite:    pkgManifest.AuthorSite,
		SourceCode:    pkgManifest.SourceCode,
		License:       pkgManifest.License,
	}

	versionResult, err := d.packageVersionsTable().Insert(version)
	if err != nil {
		return 0, err
	}

	versionId := versionResult.ID().(int64)

	err = d.fileOps.ApplyZipToFile(versionId, filePath)
	if err != nil {
		return 0, err
	}

	err = d.installedPackagesTable().Find(db.Cond{"id": installId}).Update(map[string]any{
		"active_install_id": versionId,
	})
	if err != nil {
		return 0, err
	}

	return installId, nil
}

func (d *PackageInstallOperations) GetPackage(id int64) (*dbmodels.InstalledPackage, error) {
	var pkg dbmodels.InstalledPackage
	err := d.installedPackagesTable().Find(db.Cond{"id": id}).One(&pkg)
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (d *PackageInstallOperations) DeletePackage(id int64) error {
	return d.installedPackagesTable().Find(db.Cond{"id": id}).Delete()
}

func (d *PackageInstallOperations) UpdatePackage(id int64, file string) (int64, error) {
	// Read the new file
	_, err := os.Stat(file)
	if err != nil {
		return 0, err
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return 0, err
	}

	// Create a new version
	version := &dbmodels.PackageVersion{
		Name:    filepath.Base(file),
		Slug:    "",
		Info:    "",
		Tags:    "",
		Version: "1.0.0",
	}

	result, err := d.packageVersionsTable().Insert(version)
	if err != nil {
		return 0, err
	}

	versionId := result.ID().(int64)

	// Update the package with new active install id
	err = d.installedPackagesTable().Find(db.Cond{"id": id}).Update(map[string]any{
		"active_install_id": versionId,
	})
	if err != nil {
		return 0, err
	}

	_ = content

	return versionId, nil
}

func (d *PackageInstallOperations) ListPackages() ([]dbmodels.InstalledPackage, error) {
	var packages []dbmodels.InstalledPackage
	err := d.installedPackagesTable().Find().All(&packages)
	if err != nil {
		return nil, err
	}
	return packages, nil
}

func (d *PackageInstallOperations) ListPackagesByIds(ids []int64) ([]dbmodels.InstalledPackage, error) {
	var packages []dbmodels.InstalledPackage
	err := d.installedPackagesTable().Find(db.Cond{"id IN": ids}).All(&packages)
	if err != nil {
		return nil, err
	}
	return packages, nil
}

func (d *PackageInstallOperations) ListPackageVersionByIds(ids []int64) ([]dbmodels.PackageVersion, error) {
	var versions []dbmodels.PackageVersion
	err := d.packageVersionsTable().Find(db.Cond{"id IN": ids}).All(&versions)
	if err != nil {
		return nil, err
	}
	return versions, nil
}

func (d *PackageInstallOperations) ListPackagesByInstallId(installId int64) ([]dbmodels.PackageVersion, error) {
	var versions []dbmodels.PackageVersion
	err := d.packageVersionsTable().Find(db.Cond{"id": installId}).All(&versions)
	if err != nil {
		return nil, err
	}
	return versions, nil
}

func (d *PackageInstallOperations) ListPackageVersionsByPackageId(packageId int64) ([]dbmodels.PackageVersion, error) {
	var versions []dbmodels.PackageVersion
	err := d.packageVersionsTable().Find(db.Cond{"install_id": packageId}).All(&versions)
	if err != nil {
		return nil, err
	}
	return versions, nil
}

func (d *PackageInstallOperations) GetPackageVersion(id int64) (*dbmodels.PackageVersion, error) {
	var version dbmodels.PackageVersion
	err := d.packageVersionsTable().Find(db.Cond{"id": id}).One(&version)
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (d *PackageInstallOperations) DeletePackageVersion(id int64) error {
	return d.packageVersionsTable().Find(db.Cond{"id": id}).Delete()
}

func (d *PackageInstallOperations) AddPackageVersion(installId int64, filePath string) (int64, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}

	version := &dbmodels.PackageVersion{
		Name:    filepath.Base(filePath),
		Slug:    "",
		Info:    "",
		Tags:    "",
		Version: "1.0.0",
	}

	result, err := d.packageVersionsTable().Insert(version)
	if err != nil {
		return 0, err
	}

	versionId := result.ID().(int64)
	_ = content

	return versionId, nil
}

// Private helper methods

func (d *PackageInstallOperations) installedPackagesTable() db.Collection {
	return d.db.Collection("InstalledPackages")
}

func (d *PackageInstallOperations) packageVersionsTable() db.Collection {
	return d.db.Collection("PackageVersion")
}
