package datahub

import (
	"io"
	"net/http"

	"github.com/blue-monads/turnix/backend/services/datahub/models"
	"github.com/upper/db/v4"
)

type Database interface {
	Core
	UserOps
	FileDataOps
	SpaceOps
}

type Core interface {
	Table(name string) db.Collection
	GetSession() db.Session
	RunDDL(ddl string) error
	Init() error
	Close() error
	Vender() string
}

type UserOps interface {
	AddUser(data *models.User) (int64, error)
	GetUser(id int64) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	ListUser() ([]models.User, error)
	ListUserByOwner(owner int64) ([]models.User, error)
	UpdateUser(id int64, data map[string]any) error
	DeleteUser(id int64) error
	ListUserDevice(userId int64) ([]models.UserDevice, error)
	GetDevice(id int64) (*models.UserDevice, error)
	DeleteDevice(id int64) error
	UpdateDevice(id int64, data map[string]any) error
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
