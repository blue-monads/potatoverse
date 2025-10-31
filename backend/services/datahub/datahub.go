package datahub

import (
	"io"
	"net/http"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

type Database interface {
	Core
	GetGlobalOps() GlobalOps
	GetUserOps() UserOps
	GetSpaceOps() SpaceOps
	GetSpaceKVOps() SpaceKVOps
	GetPackageInstallOps() PackageInstallOps
	GetFileOps() FileOps
	GetPackageFileOps() FileOps
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
	GetGlobalConfig(key, group string) (*dbmodels.GlobalConfig, error)
	ListGlobalConfigs(group string, offset int, limit int) ([]dbmodels.GlobalConfig, error)
	AddGlobalConfig(data *dbmodels.GlobalConfig) (int64, error)
	UpdateGlobalConfig(id int64, data map[string]any) error
	UpdateGlobalConfigByKey(key, group string, data map[string]any) error
	DeleteGlobalConfig(id int64) error
}

type UserOps interface {
	AddUserGroup(name string, info string) error
	GetUserGroup(name string) (*dbmodels.UserGroup, error)
	ListUserGroups() ([]dbmodels.UserGroup, error)
	UpdateUserGroup(name string, info string) error
	DeleteUserGroup(name string) error

	AddUser(data *dbmodels.User) (int64, error)
	GetUser(id int64) (*dbmodels.User, error)
	GetUserByEmail(email string) (*dbmodels.User, error)
	GetUserByUsername(username string) (*dbmodels.User, error)
	ListUser(offset int, limit int) ([]dbmodels.User, error)
	ListUserByOwner(owner int64) ([]dbmodels.User, error)
	UpdateUser(id int64, data map[string]any) error
	DeleteUser(id int64) error

	ListUserDevice(userId int64) ([]dbmodels.UserDevice, error)
	GetUserDevice(id int64) (*dbmodels.UserDevice, error)
	DeleteUserDevice(id int64) error
	UpdateUserDevice(id int64, data map[string]any) error

	// User Invites
	AddUserInvite(data *dbmodels.UserInvite) (int64, error)
	GetUserInvite(id int64) (*dbmodels.UserInvite, error)
	GetUserInviteByEmail(email string) (*dbmodels.UserInvite, error)
	ListUserInvites(offset int, limit int) ([]dbmodels.UserInvite, error)
	ListUserInvitesByInviter(inviterId int64) ([]dbmodels.UserInvite, error)
	UpdateUserInvite(id int64, data map[string]any) error
	DeleteUserInvite(id int64) error
}

type PackageInstallOps interface {
	InstallPackage(userId int64, repo, filePath string) (int64, error)
	GetPackage(id int64) (*dbmodels.InstalledPackage, error)
	DeletePackage(id int64) error
	UpdatePackage(id int64, file string) (int64, error)
	ListPackages() ([]dbmodels.InstalledPackage, error)
	ListPackagesByIds(ids []int64) ([]dbmodels.InstalledPackage, error)

	ListPackageVersionByIds(ids []int64) ([]dbmodels.PackageVersion, error)
	ListPackagesByInstallId(installId int64) ([]dbmodels.PackageVersion, error)
	GetPackageVersion(id int64) (*dbmodels.PackageVersion, error)
	DeletePackageVersion(id int64) error
	AddPackageVersion(installId int64, file string) (int64, error)
}

type SpaceOps interface {
	AddSpace(data *dbmodels.Space) (int64, error)
	GetSpace(id int64) (*dbmodels.Space, error)
	ListSpaces() ([]dbmodels.Space, error)
	UpdateSpace(id int64, data map[string]any) error
	RemoveSpace(id int64) error

	ListSpaceUsers(spaceId int64) ([]dbmodels.SpaceUser, error)
	AddUserToSpace(ownerId int64, userId int64, spaceId int64) error
	RemoveUserFromSpace(ownerId int64, userId int64, spaceId int64) error
	GetSpaceUserScope(userId int64, spaceId int64) (string, error)
	ListOwnSpaces(ownerId int64, spaceType string) ([]dbmodels.Space, error)
	ListThirdPartySpaces(userId int64, spaceType string) ([]dbmodels.Space, error)
	ListSpacesByPackageId(packageId int64) ([]dbmodels.Space, error)

	AddSpaceConfig(spaceId int64, uid int64, data *dbmodels.SpaceConfig) (int64, error)
	ListSpaceConfigs(spaceId int64) ([]dbmodels.SpaceConfig, error)
	GetSpaceConfig(spaceId int64, uid int64, id int64) (*dbmodels.SpaceConfig, error)
	UpdateSpaceConfig(spaceId int64, uid int64, id int64, data map[string]any) error
	RemoveSpaceConfig(spaceId int64, uid int64, id int64) error

	ListSpaceTables(spaceId int64) ([]string, error)
	ListSpaceTableColumns(spaceId int64, table string) ([]dbmodels.SpaceTableColumn, error)
	RunSpaceSQLQuery(spaceId int64, query string, data []any) ([]map[string]any, error)
	RunSpaceDDL(spaceId int64, ddl string) error
}

type SpaceKVOps interface {
	QuerySpaceKV(spaceId int64, cond map[any]any) ([]dbmodels.SpaceKV, error)
	AddSpaceKV(spaceId int64, data *dbmodels.SpaceKV) error
	GetSpaceKV(spaceId int64, group string, key string) (*dbmodels.SpaceKV, error)
	GetSpaceKVByGroup(spaceId int64, group string, offset int, limit int) ([]dbmodels.SpaceKV, error)
	RemoveSpaceKV(spaceId int64, group string, key string) error
	UpdateSpaceKV(spaceId int64, group, key string, data map[string]any) error
	UpsertSpaceKV(spaceId int64, group, key string, data map[string]any) error
}

type CreateFileRequest struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	CreatedBy int64  `json:"created_by"`
}

type FileOps interface {
	CreateFile(ownerID int64, req *CreateFileRequest, stream io.Reader) (int64, error)
	CreateFolder(ownerID int64, path string, name string, createdBy int64) (int64, error)
	GetFileContent(ownerID int64, id int64) ([]byte, error)
	GetFileContentByPath(ownerID int64, path, name string) ([]byte, error)
	GetFileMeta(id int64) (*dbmodels.FileMeta, error)
	GetFileMetaByPath(ownerID int64, path, name string) (*dbmodels.FileMeta, error)
	ListFiles(ownerID int64, path string) ([]dbmodels.FileMeta, error)
	RemoveFile(ownerID int64, id int64) error
	StreamFile(ownerID int64, id int64, w io.Writer) error
	StreamFileByPath(ownerID int64, path, name string, w io.Writer) error
	StreamFileToHTTP(ownerID int64, path, name string, w http.ResponseWriter) error
	UpdateFile(ownerID int64, id int64, stream io.Reader) error
	UpdateFileMeta(ownerID int64, id int64, data map[string]any) error

	AddFileShare(ownerID int64, fileId int64, userId int64) (string, error)
	GetSharedFile(ownerID int64, id string, w http.ResponseWriter) error
	ListFileShares(ownerID int64, fileId int64) ([]dbmodels.FileShare, error)
	RemoveFileShare(ownerID int64, userId int64, id string) error
}
