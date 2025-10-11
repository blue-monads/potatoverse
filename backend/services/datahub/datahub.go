package datahub

import (
	"io"
	"net/http"

	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/upper/db/v4"
)

type Database interface {
	Core
	GlobalOps
	UserOps
	SpaceFileOps
	SpaceOps
	SpaceKVOps
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

type PackageOps interface {
	InstallPackage(userId int64, file string) (int64, error)

	// UpgradePackage(userId, packageId int64, patchZip string) (int64, error)

	GetPackage(id int64) (*dbmodels.Package, error)
	DeletePackage(id int64) error
	UpdatePackage(id int64, data map[string]any) error

	ListPackages() ([]dbmodels.Package, error)
	ListPackagesByIds(ids []int64) ([]dbmodels.Package, error)

	ListPackageFiles(packageId int64) ([]dbmodels.PackageFile, error)
	ListPackageFilesByPath(packageId int64, path string) ([]dbmodels.PackageFile, error)
	// ListAllPackageFile(packageId int64, path string) ([]dbmodels.PackageFile, error)

	GetPackageFileMeta(packageId, id int64) (*dbmodels.PackageFile, error)
	GetPackageFileMetaByPath(packageId int64, path, name string) (*dbmodels.PackageFile, error)

	GetPackageFileStreaming(packageId, id int64, w io.Writer) error
	GetPackageFileStreamingByPath(packageId int64, path, name string, w io.Writer) error
	GetPackageFile(packageId, id int64) ([]byte, error)
	AddPackageFile(packageId int64, name string, path string, data []byte) (int64, error)
	AddPackageFileStreaming(packageId int64, name string, path string, stream io.Reader) (int64, error)
	UpdatePackageFile(packageId, id int64, data []byte) error
	UpdatePackageFileStreaming(packageId, id int64, stream io.Reader) error

	DeletePackageFile(packageId, id int64) error
}

type SpaceFileOps interface {
	StreamAddSpaceFile(spaceId int64, uid int64, path string, name string, stream io.Reader) (id int64, err error)
	AddSpaceFolder(spaceId int64, uid int64, path string, name string) (int64, error)

	GetSpaceFileMetaByPath(spaceId int64, path string) (*dbmodels.File, error)
	GetSpaceFileMetaByPathAndName(spaceId int64, path string, name string) (*dbmodels.File, error)
	GetSpaceFileMetaById(id int64) (*dbmodels.File, error)
	GetSpaceFile(spaceId int64, id int64) ([]byte, error)
	StreamGetSpaceFile(spaceId int64, uid int64, id int64, w http.ResponseWriter) error
	StreamGetSpaceFileByPath(spaceId int64, uid int64, path string, name string, w http.ResponseWriter) error

	RemoveSpaceFile(spaceId, id int64) error
	UpdateSpaceFile(spaceId, id int64, data map[string]any) error
	StreamUpdateSpaceFile(spaceId, id int64, stream io.Reader) (int64, error)
	ListSpaceFiles(spaceId int64, path string) ([]dbmodels.File, error)

	// File Shares

	AddFileShare(fileId int64, userId int64, spaceId int64) (string, error)
	GetSharedFile(id string, w http.ResponseWriter) error
	ListFileShares(fileId int64) ([]dbmodels.FileShare, error)
	RemoveFileShare(userId int64, id string) error
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
