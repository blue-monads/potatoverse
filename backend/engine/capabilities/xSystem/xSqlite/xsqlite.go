package xsqlite

import (
	"database/sql"
	"sync"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

type SqliteCapability struct {
	db    *sql.DB
	txns  map[string]*sql.Tx
	txnMu sync.Mutex
}

func (c *SqliteCapability) Handle(ctx *gin.Context) {}

func (c *SqliteCapability) ListActions() ([]string, error) {
	return []string{
		"execute",
		"query",
		"query_row",
		"create_txn",
		"rollback_txn",
		"commit_txn",
		"txn_execute",
		"txn_query",
		"txn_query_row",
	}, nil
}

func (c *SqliteCapability) Reload(model *dbmodels.SpaceCapability) (xcapability.Capability, error) {
	return c, nil
}

func (c *SqliteCapability) Close() error {
	c.txnMu.Lock()
	for id, tx := range c.txns {
		tx.Rollback()
		delete(c.txns, id)
	}
	c.txnMu.Unlock()

	return c.db.Close()
}
