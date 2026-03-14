package xsqlite

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability"
	"github.com/gin-gonic/gin"
)

var (
	Name         = "xSqlite"
	Icon         = `<i class="fa-solid fa-database"></i>`
	OptionFields = []xcapability.CapabilityOptionField{}
)

func init() {
	b := xcapability.CapabilityBuilderFactory{
		Builder: func(app any) (xcapability.CapabilityBuilder, error) {
			return &SqliteBuilder{
				dbs: make(map[string]*SqliteCapability),
			}, nil
		},
		Name:         Name,
		Icon:         Icon,
		OptionFields: OptionFields,
	}

	registry.RegisterCapability(Name, b)
	registry.RegisterCapability("xsqlite", b)
}

type SqliteBuilder struct {
	dbs  map[string]*SqliteCapability
	lock sync.Mutex
}

func (b *SqliteBuilder) Build(handle xcapability.XCapabilityHandle) (xcapability.Capability, error) {
	model := handle.GetModel()
	dbKey := fmt.Sprintf("cap-%d", model.ID)

	sdb, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to open in-memory sqlite: %w", err)
	}

	sdb.SetMaxOpenConns(1)

	cap := &SqliteCapability{
		db:    sdb,
		txns:  make(map[string]*sql.Tx),
		txnMu: sync.Mutex{},
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	if existing, ok := b.dbs[dbKey]; ok {
		existing.Close()
	}
	b.dbs[dbKey] = cap

	return cap, nil
}

func (b *SqliteBuilder) Serve(ctx *gin.Context) {}

func (b *SqliteBuilder) Name() string {
	return Name
}

func (b *SqliteBuilder) GetDebugData() map[string]any {
	b.lock.Lock()
	defer b.lock.Unlock()

	info := make(map[string]any)
	for k, v := range b.dbs {
		v.txnMu.Lock()
		info[k] = map[string]any{
			"active_txns": len(v.txns),
		}
		v.txnMu.Unlock()
	}

	return map[string]any{
		"databases": info,
	}
}
