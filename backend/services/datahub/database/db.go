package database

import (
	"database/sql"
	_ "embed"
	"errors"
	"log"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/sqlite"
)

//go:embed schema.sql
var schema string

type DB struct {
	sess                 db.Session
	externalFileMode     bool
	minFileMultiPartSize int64
}

const (
	ScopeOwner = "owner"
)

var (
	ErrUserNoScope = errors.New("err: user doesnot have required scope")
)

func NewDB(file string) (*DB, error) {

	var settings = sqlite.ConnectionURL{
		Database: file,
	}

	sess, err := sqlite.Open(settings)
	if err != nil {
		log.Fatalf("db.Open(): %q\n", err)
	}

	exists, _ := sess.Collection("Users").Exists()

	if !exists {
		driver := sess.Driver().(*sql.DB)

		_, err = driver.Exec(schema)
		if err != nil {
			sess.Close()
			return nil, err
		}

	}

	return &DB{
		sess:                 sess,
		externalFileMode:     false,
		minFileMultiPartSize: 1024 * 1024 * 8,
	}, nil
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

func (db *DB) GetSession() db.Session {
	return db.sess
}

func (db *DB) Table(name string) db.Collection {
	return db.sess.Collection(name)
}
