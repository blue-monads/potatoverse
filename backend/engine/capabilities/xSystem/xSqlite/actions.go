package xsqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/blue-monads/potatoverse/backend/utils/libx/dbutils"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
)

var txnCounter atomic.Int64

var okResult = struct {
	Success bool `json:"success"`
}{
	Success: true,
}

func (c *SqliteCapability) Execute(name string, params lazydata.LazyData) (any, error) {
	switch name {
	case "execute":
		return c.execSQL(params)
	case "query":
		return c.querySQL(params)
	case "query_row":
		return c.queryRowSQL(params)
	case "create_txn":
		return c.createTxn(params)
	case "rollback_txn":
		return c.rollbackTxn(params)
	case "commit_txn":
		return c.commitTxn(params)
	case "txn_execute":
		return c.txnExecSQL(params)
	case "txn_query":
		return c.txnQuerySQL(params)
	case "txn_query_row":
		return c.txnQueryRowSQL(params)
	default:
		return nil, errors.New("unknown action: " + name)
	}
}

type sqlParams struct {
	Query string `json:"query"`
	Args  []any  `json:"args"`
}

type txnSqlParams struct {
	TxnId string `json:"txn_id"`
	Query string `json:"query"`
	Args  []any  `json:"args"`
}

type execResult struct {
	LastInsertId int64 `json:"last_insert_id"`
	RowsAffected int64 `json:"rows_affected"`
}

// execSQL runs INSERT, UPDATE, DELETE, CREATE TABLE, etc.
func (c *SqliteCapability) execSQL(params lazydata.LazyData) (any, error) {
	var p sqlParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	result, err := c.db.Exec(p.Query, p.Args...)
	if err != nil {
		return nil, err
	}

	return toExecResult(result)
}

func (c *SqliteCapability) querySQL(params lazydata.LazyData) (any, error) {
	var p sqlParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	rows, err := c.db.Query(p.Query, p.Args...)
	if err != nil {
		return nil, err
	}

	results, err := dbutils.SelectScan(rows)
	if err != nil {
		return nil, err
	}
	if results == nil {
		return []map[string]any{}, nil
	}
	return results, nil
}

func (c *SqliteCapability) queryRowSQL(params lazydata.LazyData) (any, error) {
	var p sqlParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	rows, err := c.db.Query(p.Query, p.Args...)
	if err != nil {
		return nil, err
	}

	result, err := dbutils.GetScan(rows)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return result, err
}

func (c *SqliteCapability) createTxn(_ lazydata.LazyData) (any, error) {
	tx, err := c.db.Begin()
	if err != nil {
		return nil, err
	}

	txnId := fmt.Sprintf("txn_%d", txnCounter.Add(1))

	c.txnMu.Lock()
	c.txns[txnId] = tx
	c.txnMu.Unlock()

	return map[string]string{"txn_id": txnId}, nil
}

func (c *SqliteCapability) getTxn(txnId string) (*sql.Tx, error) {
	c.txnMu.Lock()
	defer c.txnMu.Unlock()

	tx, ok := c.txns[txnId]
	if !ok {
		return nil, fmt.Errorf("transaction not found: %s", txnId)
	}
	return tx, nil
}

func (c *SqliteCapability) removeTxn(txnId string) {
	c.txnMu.Lock()
	delete(c.txns, txnId)
	c.txnMu.Unlock()
}

func (c *SqliteCapability) rollbackTxn(params lazydata.LazyData) (any, error) {
	txnId := params.GetFieldAsString("txn_id")
	if txnId == "" {
		return nil, errors.New("txn_id is required")
	}

	tx, err := c.getTxn(txnId)
	if err != nil {
		return nil, err
	}

	err = tx.Rollback()
	c.removeTxn(txnId)
	if err != nil {
		return nil, err
	}

	return okResult, nil
}

func (c *SqliteCapability) commitTxn(params lazydata.LazyData) (any, error) {
	txnId := params.GetFieldAsString("txn_id")
	if txnId == "" {
		return nil, errors.New("txn_id is required")
	}

	tx, err := c.getTxn(txnId)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	c.removeTxn(txnId)
	if err != nil {
		return nil, err
	}

	return okResult, nil
}

func (c *SqliteCapability) txnExecSQL(params lazydata.LazyData) (any, error) {
	var p txnSqlParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	tx, err := c.getTxn(p.TxnId)
	if err != nil {
		return nil, err
	}

	result, err := tx.Exec(p.Query, p.Args...)
	if err != nil {
		return nil, err
	}

	return toExecResult(result)
}

func (c *SqliteCapability) txnQuerySQL(params lazydata.LazyData) (any, error) {
	var p txnSqlParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	tx, err := c.getTxn(p.TxnId)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(p.Query, p.Args...)
	if err != nil {
		return nil, err
	}

	results, err := dbutils.SelectScan(rows)
	if err != nil {
		return nil, err
	}
	if results == nil {
		return []map[string]any{}, nil
	}
	return results, nil
}

func (c *SqliteCapability) txnQueryRowSQL(params lazydata.LazyData) (any, error) {
	var p txnSqlParams
	if err := params.AsJson(&p); err != nil {
		return nil, err
	}

	tx, err := c.getTxn(p.TxnId)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(p.Query, p.Args...)
	if err != nil {
		return nil, err
	}

	result, err := dbutils.GetScan(rows)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return result, err
}

func toExecResult(result sql.Result) (*execResult, error) {
	lastId, _ := result.LastInsertId()
	affected, _ := result.RowsAffected()

	return &execResult{
		LastInsertId: lastId,
		RowsAffected: affected,
	}, nil
}
