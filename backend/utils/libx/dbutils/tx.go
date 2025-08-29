package dbutils

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/upper/db/v4"
)

var (
	ErrTxNotFound = errors.New("Tx: not found")
)

type TxManager struct {
	openTxns         map[uint32]*TxnWrapped
	atomicTxnCounter uint32
	txBuilder        func(sqlTx *sql.Tx) (Tx, error)
	defaultTimeOut   time.Duration
}

func NewTxMgr(fn func(sqlTx *sql.Tx) (Tx, error)) TxManager {
	return TxManager{
		openTxns:         make(map[uint32]*TxnWrapped),
		atomicTxnCounter: 2, // 0 and 1 are taken
		txBuilder:        fn,
		defaultTimeOut:   time.Millisecond * time.Duration(10),
	}
}

// bcz sqlbuilder.Tx is internal/private package
// cannot import it
type Tx interface {
	db.Session
	Commit() error
	Rollback() error
}

func (d *TxManager) NewTxn(sess db.Session, timeout int) (uint32, error) {
	driver := sess.Driver().(*sql.DB)
	txid := atomic.AddUint32(&d.atomicTxnCounter, 1)

	mxtime := d.defaultTimeOut

	if timeout != 0 {
		mxtime = time.Millisecond * time.Duration(timeout)
	}

	ctx, ctxFn := context.WithTimeout(context.Background(), mxtime)
	tx, err := driver.BeginTx(ctx, nil)
	if err != nil {
		ctxFn()
		return 0, err
	}

	txn, err := d.txBuilder(tx)
	if err != nil {
		ctxFn()
		return 0, err
	}

	wtx := &TxnWrapped{
		tx:          txn,
		mlock:       sync.Mutex{},
		cancelCtxFn: ctxFn,
	}

	d.openTxns[txid] = wtx

	time.AfterFunc(time.Millisecond*time.Duration(timeout), func() {
		delete(d.openTxns, txid)
		wtx.Close()
	})

	return txid, nil
}

func (d *TxManager) CommitTx(txid uint32) error {
	wtxn, ok := d.openTxns[txid]
	if !ok {
		return errors.New("not found")
	}
	wtxn.mlock.Lock()
	defer wtxn.mlock.Unlock()
	return wtxn.tx.Commit()
}

func (d *TxManager) RollbackTx(txid uint32) error {
	wtxn, ok := d.openTxns[txid]
	if !ok {
		return errors.New("not found")
	}
	wtxn.mlock.Lock()
	defer wtxn.mlock.Unlock()
	return wtxn.tx.Rollback()
}

func (d *TxManager) clearTxn(txids ...uint32) {
	for _, txid := range txids {
		wtxn, ok := d.openTxns[txid]
		if !ok {
			return
		}
		wtxn.Close()
		delete(d.openTxns, txid)
	}
}

// TxOr runs in txn or directly db connection
func (d *TxManager) TxOr(txid uint32, dbsess db.Session, fn func(sess db.Session) error) error {
	if txid == 0 {
		return fn(dbsess)
	}

	if txid == 1 {
		_txid, err := d.NewTxn(dbsess, 0)
		if err != nil {
			return err
		}
		txid = _txid
	}

	wtxn, ok := d.openTxns[txid]
	if !ok {
		return ErrTxNotFound
	}

	wtxn.mlock.Lock()
	defer wtxn.mlock.Unlock()
	return fn(wtxn.tx)
}

type TxnWrapped struct {
	tx          Tx
	mlock       sync.Mutex
	cancelCtxFn func()
}

func (t *TxnWrapped) Close() {
	t.mlock.Lock()
	defer t.mlock.Unlock()
	t.cancelCtxFn()
}
