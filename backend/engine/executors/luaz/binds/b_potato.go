package binds

import (
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

const (
	luaKVModuleTypeName   = "potato.kv"
	luaCapModuleTypeName  = "potato.cap"
	luaDBModuleTypeName   = "potato.db"
	luaTxnModuleTypeName  = "potato.txn"
	luaCoreModuleTypeName = "potato.core"
)

type luaCapModule struct {
	app          xtypes.App
	installId    int64
	spaceId      int64
	capabilities xtypes.CapabilityHub
}

type luaDBModule struct {
	installId int64
	db        datahub.DBLowCoreOps
}

type luaTxnModule struct {
	installId int64
	txn       datahub.DBLowTxnOps
}

type luaCoreModule struct {
	app       xtypes.App
	installId int64
	spaceId   int64
	engine    xtypes.Engine
	sig       *signer.Signer
}

func PotatoModule(app xtypes.App, installId int64, spaceId int64) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		// Register type metatables
		registerKVModuleType(L)
		registerCapModuleType(L)
		registerDBModuleType(L)
		registerTxnModuleType(L)
		registerCoreModuleType(L)

		// Create main potato table
		potatoTable := L.NewTable()

		// Create sub-modules as userdata
		kvModule := newKVModule(L, app, installId)
		capModule := newCapModule(L, app, installId, spaceId)
		dbModule := newDBModule(L, app, installId)
		coreModule := newCoreModule(L, app, installId, spaceId)

		// Set sub-modules on main table
		potatoTable.RawSetString("kv", kvModule)
		potatoTable.RawSetString("cap", capModule)
		potatoTable.RawSetString("db", dbModule)
		potatoTable.RawSetString("core", coreModule)

		L.Push(potatoTable)
		return 1
	}
}
