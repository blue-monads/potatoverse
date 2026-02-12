package database

import (
	"bytes"
	"database/sql"
	_ "embed"
	"errors"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database/event"
	fileops "github.com/blue-monads/potatoverse/backend/services/datahub/database/file"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database/global"
	ppackage "github.com/blue-monads/potatoverse/backend/services/datahub/database/ppackage"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database/schema"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database/space"
	"github.com/blue-monads/potatoverse/backend/services/datahub/database/user"
	"github.com/blue-monads/potatoverse/backend/services/datahub/lazysyncer"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/upper/db/v4"
	upperdb "github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

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

	lazySyncer *lazysyncer.LazySyncer
}

const (
	ScopeOwner = "owner"
)

const (
	CDC_ENABLED = false
)

var (
	ErrUserNoScope = errors.New("err: user doesnot have required scope")
)

func NewDB(file string, logger *slog.Logger) (*DB, error) {

	if runtime.GOOS == "windows" {
		directPath := os.Getenv("DIRECT_DB_PATH")

		if directPath != "" {
			file = directPath
		}
	}

	qq.Println("@final_path", file)

	var settings = sqlite.ConnectionURL{
		Database: file,
		Options: map[string]string{
			"_journal_mode": "WAL",
			"_busy_timeout": "10000", // 10 second busy timeout
		},
	}

	if logger != nil {
		logger.Info("opening sqlite", "file", file)
	}

	sess, err := sqlite.Open(settings)
	if err != nil {
		logger.Error("sqlite.Open() failed", "err", err)
		return nil, err
	}

	return fromSqlHandle(sess, logger)
}

func AutoMigrate(sess upperdb.Session) error {

	exists, _ := sess.Collection("Users").Exists()

	if !exists {
		driver := sess.Driver().(*sql.DB)

		buf := bytes.Buffer{}

		pschema := strings.Replace(fileops.FileSchemaSQL, "FileMeta", "PFileMeta", 1)
		pschema = strings.Replace(pschema, "FileBlob", "PFileBlob", 1)
		pschema = strings.Replace(pschema, "FileShares", "PFileShares", 1)

		schema := schema.Get()

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

func FromSqlHandle(sdb *sql.DB, logger *slog.Logger) (*DB, error) {
	sess, err := sqlite.New(sdb)
	if err != nil {
		return nil, err
	}

	if logger == nil {
		logger = slog.Default()
	}

	return fromSqlHandle(sess, logger)
}

func fromSqlHandle(sess upperdb.Session, logger *slog.Logger) (*DB, error) {

	sdb := sess.Driver().(*sql.DB)

	_, err := sdb.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		return nil, err
	}

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
	var lazySyncer *lazysyncer.LazySyncer

	if CDC_ENABLED {
		lazySyncer = lazysyncer.New(lazysyncer.Options{
			DbSession:     sess,
			IsSelfEnabled: CDC_ENABLED,
			Buddies:       []string{},
			BasePath:      "./buddies",
			Logger:        logger,
		})

	}

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
		lazySyncer:           lazySyncer,
	}, nil
}

func (db *DB) Init(transport datahub.BuddyTransport) error {

	debugInfo, err := db.GetDbStates()
	if err != nil {
		return err
	}

	qq.Println("@db_debug_info", debugInfo)

	if db.lazySyncer != nil {
		if err := db.lazySyncer.Start(transport); err != nil {
			return err
		}
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

func (db *DB) IsEmptyRowsError(err error) bool {
	return errors.Is(err, upperdb.ErrNoMoreRows)
}

func (db *DB) GetDbStates() (map[string]any, error) {
	driver := db.sess.Driver().(*sql.DB)
	return GetDbStates(driver)
}
