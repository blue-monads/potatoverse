package datahub

import (
	"io"
	"net/http"

	"github.com/blue-monads/turnix/backend/services/datahub/models"
	"github.com/upper/db/v4"
)

type Database interface {
	Core
	GlobalOps
	UserOps
	FileDataOps
	SpaceOps
	PackageOps
}

type Core interface {
	Table(name string) db.Collection
	GetSession() db.Session
	RunDDL(ddl string) error
	Init() error
	Close() error
	Vender() string
	HasTable(name string) (bool, error)

	IsEmptyRowsError(err error) bool
}

type GlobalOps interface {
	GetGlobalConfig(key, group string) (*models.GlobalConfig, error)
	ListGlobalConfigs(group string, offset int, limit int) ([]models.GlobalConfig, error)
	AddGlobalConfig(data *models.GlobalConfig) (int64, error)
	UpdateGlobalConfig(id int64, data map[string]any) error
	UpdateGlobalConfigByKey(key, group string, data map[string]any) error
	DeleteGlobalConfig(id int64) error
}

type UserOps interface {
	AddUser(data *models.User) (int64, error)
	GetUser(id int64) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	ListUser(offset int, limit int) ([]models.User, error)
	ListUserByOwner(owner int64) ([]models.User, error)
	UpdateUser(id int64, data map[string]any) error
	DeleteUser(id int64) error

	ListUserDevice(userId int64) ([]models.UserDevice, error)
	GetUserDevice(id int64) (*models.UserDevice, error)
	DeleteUserDevice(id int64) error
	UpdateUserDevice(id int64, data map[string]any) error
}

type PackageOps interface {
	InstallPackage(userId int64, file string) (int64, error)
	GetPackage(id int64) (*models.Package, error)
	DeletePackage(id int64) error
	UpdatePackage(id int64, data map[string]any) error

	ListPackages() ([]models.Package, error)
	ListPackagesByIds(ids []int64) ([]models.Package, error)

	ListPackageFiles(packageId int64) ([]models.PackageFile, error)
	GetPackageFileMeta(packageId, id int64) (*models.PackageFile, error)
	GetPackageFileMetaByPath(packageId int64, path, name string) (*models.PackageFile, error)

	GetPackageFileStreaming(packageId, id int64, w io.Writer) error
	GetPackageFile(packageId, id int64) ([]byte, error)
	AddPackageFile(packageId int64, name string, path string, data []byte) (int64, error)
	AddPackageFileStreaming(packageId int64, name string, path string, stream io.Reader) (int64, error)
	UpdatePackageFile(packageId, id int64, data []byte) error
	UpdatePackageFileStreaming(packageId, id int64, stream io.Reader) error

	DeletePackageFile(packageId, id int64) error
}

type FileDataOps interface {
	AddFileShare(fileId int64, userId int64, spaceId int64) (string, error)
	AddFileStreaming(file *models.File, stream io.Reader) (id int64, err error)
	AddFolder(spaceId int64, uid int64, path string, name string) (int64, error)
	GetFileBlobStreaming(id int64, w http.ResponseWriter) error
	GetFileMeta(id int64) (*models.File, error)
	GetSharedFile(id string, w http.ResponseWriter) error
	ListFileShares(fileId int64) ([]models.FileShare, error)
	ListFilesBySpace(spaceId int64, path string) ([]models.File, error)
	ListFilesByUser(uid int64, path string) ([]models.File, error)
	RemoveFileShare(userId int64, id string) error
	RemoveFile(id int64) error
	UpdateFile(id int64, data map[string]any) error
	UpdateFileStreaming(file *models.File, stream io.Reader) (int64, error)
}

type SpaceOps interface {
	AddSpace(data *models.Space) (int64, error)
	GetSpace(id int64) (*models.Space, error)
	ListSpaces() ([]models.Space, error)
	UpdateSpace(id int64, data map[string]any) error
	RemoveSpace(id int64) error

	ListSpaceUsers(spaceId int64) ([]models.SpaceUser, error)
	AddUserToSpace(ownerId int64, userId int64, spaceId int64) error
	RemoveUserFromSpace(ownerId int64, userId int64, spaceId int64) error
	GetSpaceUserScope(userId int64, spaceId int64) (string, error)
	ListOwnSpaces(ownerId int64, spaceType string) ([]models.Space, error)
	ListThirdPartySpaces(userId int64, spaceType string) ([]models.Space, error)

	AddSpaceConfig(spaceId int64, uid int64, data *models.SpaceConfig) (int64, error)
	ListSpaceConfigs(spaceId int64) ([]models.SpaceConfig, error)
	GetSpaceConfig(spaceId int64, uid int64, id int64) (*models.SpaceConfig, error)
	UpdateSpaceConfig(spaceId int64, uid int64, id int64, data map[string]any) error
	RemoveSpaceConfig(spaceId int64, uid int64, id int64) error

	ListSpaceTables(spaceId int64) ([]string, error)
	ListSpaceTableColumns(spaceId int64, table string) ([]models.SpaceTableColumn, error)
	RunSpaceSQLQuery(spaceId int64, query string, data []any) ([]map[string]any, error)
	RunSpaceDDL(spaceId int64, ddl string) error
}
