package binds

import (
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/utils/luaplus"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

type SpaceKVQuery struct {
	Group        string      `json:"group"`
	Cond         map[any]any `json:"cond"`
	Offset       int         `json:"offset"`
	Limit        int         `json:"limit"`
	IncludeValue bool        `json:"include_value"`
}

type luaKVModule struct {
	app       xtypes.App
	installId int64
	db        datahub.SpaceKVOps
}

// KV Module
func registerKVModuleType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaKVModuleTypeName)
	L.SetField(mt, "__index", L.NewFunction(kvModuleIndex))
}

func newKVModule(L *lua.LState, app xtypes.App, installId int64) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = &luaKVModule{
		app:       app,
		installId: installId,
		db:        app.Database().GetSpaceKVOps(),
	}
	L.SetMetatable(ud, L.GetTypeMetatable(luaKVModuleTypeName))
	return ud
}

func checkKVModule(L *lua.LState) *luaKVModule {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*luaKVModule); ok {
		return v
	}
	L.ArgError(1, luaKVModuleTypeName+" expected")
	return nil
}

func kvModuleIndex(L *lua.LState) int {
	mod := checkKVModule(L)
	method := L.CheckString(2)

	switch method {
	case "query":
		return kvQuery(mod, L)
	case "add":
		return kvAdd(mod, L)
	case "get":
		return kvGet(mod, L)
	case "get_by_group":
		return kvGetByGroup(mod, L)
	case "remove":
		return kvRemove(mod, L)
	case "update":
		return kvUpdate(mod, L)
	case "upsert":
		return kvUpsert(mod, L)
	}

	return 0
}

func kvQuery(mod *luaKVModule, L *lua.LState) int {
	query := &SpaceKVQuery{}
	err := luaplus.MapToStruct(L, L.CheckTable(1), query)
	if err != nil {
		return pushError(L, err)
	}

	if query.Group != "" {
		if query.Cond == nil {
			query.Cond = make(map[any]any)
		}
		query.Cond["group"] = query.Group
	}

	qq.Println("query.Cond:", query)

	var datas []dbmodels.SpaceKV
	if query.IncludeValue {
		datas, err = mod.db.QueryWithValueSpaceKV(mod.installId, query.Cond, query.Offset, query.Limit)
	} else {
		datas, err = mod.db.QuerySpaceKV(mod.installId, query.Cond, query.Offset, query.Limit)
	}
	if err != nil {
		return pushError(L, err)
	}

	result := L.NewTable()
	for _, data := range datas {
		luaTable, err := luaplus.StructToTable(L, data)
		if err != nil {
			return pushError(L, err)
		}
		result.Append(luaTable)
	}

	L.Push(result)
	return 1
}

func kvAdd(mod *luaKVModule, L *lua.LState) int {
	dataStruct := &dbmodels.SpaceKV{}

	luaTable, err := luaplus.StructToTable(L, dataStruct)
	if err != nil {
		return pushError(L, err)
	}

	err = mod.db.AddSpaceKV(mod.installId, dataStruct)
	if err != nil {
		return pushError(L, err)
	}
	L.Push(luaTable)
	return 1
}

func kvGet(mod *luaKVModule, L *lua.LState) int {
	group := L.CheckString(1)
	key := L.CheckString(2)
	data, err := mod.db.GetSpaceKV(mod.installId, group, key)
	if err != nil {
		return pushError(L, err)
	}

	luaTable, err := luaplus.StructToTable(L, data)
	if err != nil {
		return pushError(L, err)
	}

	L.Push(luaTable)
	return 1
}

func kvGetByGroup(mod *luaKVModule, L *lua.LState) int {
	group := L.CheckString(1)
	offset := L.CheckInt(2)
	limit := L.CheckInt(3)
	datas, err := mod.db.GetSpaceKVByGroup(mod.installId, group, offset, limit)
	if err != nil {
		return pushError(L, err)
	}

	result := L.NewTable()
	for _, data := range datas {
		luaTable, err := luaplus.StructToTable(L, data)
		if err != nil {
			return pushError(L, err)
		}
		result.Append(luaTable)
	}

	L.Push(result)
	return 1
}

func kvRemove(mod *luaKVModule, L *lua.LState) int {
	group := L.CheckString(1)
	key := L.CheckString(2)
	err := mod.db.RemoveSpaceKV(mod.installId, group, key)
	if err != nil {
		return pushError(L, err)
	}
	L.Push(lua.LNil)
	return 1
}

func kvUpdate(mod *luaKVModule, L *lua.LState) int {
	group := L.CheckString(1)
	key := L.CheckString(2)
	data := L.CheckTable(3)
	dataMap := luaplus.TableToMap(L, data)

	err := mod.db.UpdateSpaceKV(mod.installId, group, key, dataMap)
	if err != nil {
		return pushError(L, err)
	}
	L.Push(lua.LNil)
	return 1
}

func kvUpsert(mod *luaKVModule, L *lua.LState) int {
	group := L.CheckString(1)
	key := L.CheckString(2)
	data := L.CheckTable(3)
	dataMap := luaplus.TableToMap(L, data)
	err := mod.db.UpsertSpaceKV(mod.installId, group, key, dataMap)
	if err != nil {
		return pushError(L, err)
	}

	L.Push(lua.LNil)
	return 1
}
