package database

import (
	"bytes"
	"database/sql"
	_ "embed"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/cdc"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database/event"
	fileops "github.com/blue-monads/potatoverse/backend/services/datahub/database/file"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database/global"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database/low"
	ppackage "github.com/blue-monads/potatoverse/backend/services/datahub/database/ppackage"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database/space"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database/user"
	"github.com/upper/db/v4"
	upperdb "github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

//go:embed schema.sql
var schema string

type DB struct {
	sess                 upperdb.Session
	minFileMultiPartSize int64

	userOps           *user.UserOperations
	globalOps         *global.GlobalOperations
	spaceOps          *space.SpaceOperations
	fileOps           *fileops.FileOperations
	packageFileOps    *fileops.FileOperations
	packageInstallOps *ppackage.PackageInstallOperations
	eventOps          *event.EventOperations

	cdcSyncer *cdc.CDCSyncer
}

const (
	ScopeOwner = "owner"
)

var (
	ErrUserNoScope = errors.New("err: user doesnot have required scope")
)

func NewDB(file string, logger *slog.Logger) (*DB, error) {

	var settings = sqlite.ConnectionURL{
		Database: file,
	}

	sess, err := sqlite.Open(settings)
	if err != nil {
		logger.Error("sqlite.Open() failed", "err", err)
		return nil, err
	}

	return fromSqlHandle(sess)
}

func AutoMigrate(sess upperdb.Session) error {

	exists, _ := sess.Collection("Users").Exists()

	if !exists {
		driver := sess.Driver().(*sql.DB)

		buf := bytes.Buffer{}

		pschema := strings.Replace(fileops.FileSchemaSQL, "FileMeta", "PFileMeta", 1)
		pschema = strings.Replace(pschema, "FileBlob", "PFileBlob", 1)
		pschema = strings.Replace(pschema, "FileShares", "PFileShares", 1)

		buf.WriteString(schema)
		buf.WriteString("\n")
		buf.WriteString(fileops.FileSchemaSQL)
		buf.WriteString("\n")
		buf.WriteString(pschema)

		fileSchema := buf.String()

		// os.WriteFile("file_schema_patched.sql", []byte(fileSchema), 0644)

		_, err := driver.Exec(fileSchema)
		if err != nil {
			sess.Close()
			return err
		}
	}

	return nil
}

func FromSqlHandle(sdb *sql.DB) (*DB, error) {
	sess, err := sqlite.New(sdb)
	if err != nil {
		return nil, err
	}

	return fromSqlHandle(sess)
}

func fromSqlHandle(sess upperdb.Session) (*DB, error) {

	// Initialize operations
	globalOps := global.NewGlobalOperations(sess)
	spaceOps := space.NewSpaceOperations(sess)

	fileOps := fileops.NewFileOperations(fileops.Options{
		DbSess:           sess,
		MinMultiPartSize: 1024 * 1024 * 8,
		StoreType:        fileops.StoreTypeMultipart,
	})

	packageFileOps := fileops.NewFileOperations(fileops.Options{
		DbSess:           sess,
		MinMultiPartSize: 1024 * 1024 * 8,
		Prefix:           "P",
		StoreType:        fileops.StoreTypeMultipart,
	})

	packageInstallOps := ppackage.NewPackageInstallOperations(sess, packageFileOps)
	eventOps := event.NewEventOperations(sess)

	if err := AutoMigrate(sess); err != nil {
		return nil, err
	}

	cdcSyncer := cdc.NewCDCSyncer(sess, CDC_ENABLED)

	return &DB{
		sess:                 sess,
		minFileMultiPartSize: 1024 * 1024 * 8,
		userOps:              user.NewUserOperations(sess),
		globalOps:            globalOps,
		spaceOps:             spaceOps,
		fileOps:              fileOps,
		packageFileOps:       packageFileOps,
		packageInstallOps:    packageInstallOps,
		eventOps:             eventOps,
		cdcSyncer:            cdcSyncer,
	}, nil
}

const (
	CDC_ENABLED = true
)

func (db *DB) Init() error {

	if err := db.cdcSyncer.Start(); err != nil {
		return err
	}

	return nil
}

func (db *DB) Close() error {
	return db.sess.Close()
}

func (db *DB) Vender() string {
	return "sqlite"
}

func (db *DB) RunDDL(ddl string) error {
	driver := db.sess.Driver().(*sql.DB)

	_, err := driver.Exec(ddl)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetSession() upperdb.Session {
	return db.sess
}

func (db *DB) HasTable(name string) (bool, error) {
	table := db.Table(name)
	if table == nil {
		return false, nil
	}

	exists, err := table.Exists()
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (db *DB) Table(name string) db.Collection {
	return db.sess.Collection(name)
}

const ErrText = "upper: no more rows in this result set"

func (db *DB) IsEmptyRowsError(err error) bool {
	return err.Error() == ErrText
}

func (db *DB) GetGlobalOps() datahub.GlobalOps {
	return db.globalOps
}

func (db *DB) GetUserOps() datahub.UserOps {
	return db.userOps
}

func (db *DB) GetSpaceOps() datahub.SpaceOps {
	return db.spaceOps
}

func (db *DB) GetSpaceKVOps() datahub.SpaceKVOps {
	return db.spaceOps
}

func (db *DB) GetPackageInstallOps() datahub.PackageInstallOps {
	return db.packageInstallOps
}

func (db *DB) GetFileOps() datahub.FileOps {
	return db.fileOps
}

func (db *DB) GetPackageFileOps() datahub.FileOps {
	return db.packageFileOps
}

func (db *DB) GetLowDBOps(ownerType string, ownerID string) datahub.DBLowOps {
	return low.NewLowDB(db.sess, ownerType, ownerID)
}

func (db *DB) GetLowPackageDBOps(installId int64) datahub.DBLowOps {
	return low.NewLowDB(db.sess, "P", strconv.FormatInt(installId, 10))
}

func (db *DB) GetLowCapabilityDBOps(capabilityId int64) datahub.DBLowOps {
	return low.NewLowDB(db.sess, "C", strconv.FormatInt(capabilityId, 10))
}

func (db *DB) GetMQSynk() datahub.MQSynk {
	return db.eventOps
}
