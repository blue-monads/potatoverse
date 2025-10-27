package database

import (
	"database/sql"
	_ "embed"
	"errors"
	"log/slog"

	"github.com/blue-monads/turnix/backend/services/datahub/database/operations/user"
	"github.com/upper/db/v4"
	upperdb "github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

//go:embed schema.sql
var schema string

type DB struct {
	sess                 upperdb.Session
	externalFileMode     bool
	minFileMultiPartSize int64

	userOps *user.UserOperations
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

	if err := AutoMigrate(sess); err != nil {
		logger.Error("AutoMigrate() failed", "err", err)
		return nil, err
	}

	return &DB{
		sess:                 sess,
		externalFileMode:     false,
		minFileMultiPartSize: 1024 * 1024 * 8,
		userOps:              user.NewUserOperations(sess),
	}, nil
}

func AutoMigrate(sess upperdb.Session) error {

	exists, _ := sess.Collection("Users").Exists()

	if !exists {
		driver := sess.Driver().(*sql.DB)
		_, err := driver.Exec(schema)
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

	db := &DB{
		sess:                 sess,
		externalFileMode:     false,
		minFileMultiPartSize: 1024 * 1024 * 8,
	}

	if err := AutoMigrate(sess); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) Init() error {

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
