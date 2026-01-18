package binds2

import (
	"github.com/blue-monads/potatoverse/backend/engine/executors"
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/utils/luaplus"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

type SpaceKVQuery struct {
	Group        string      `json:"group"`
	Cond         map[any]any `json:"cond"`
	Offset       int         `json:"offset"`
	Limit        int         `json:"limit"`
	IncludeValue bool        `json:"include_value"`
}

func KVBindable(app xtypes.App) map[string]lua.LGFunction {
	db := app.Database().GetSpaceKVOps()

	return map[string]lua.LGFunction{
		"kv_query": func(L *lua.LState) int {
			return kvQuery(GetExecState(L), db, L)
		},
		"kv_add": func(L *lua.LState) int {
			return kvAdd(GetExecState(L), db, L)
		},
		"kv_get": func(L *lua.LState) int {
			return kvGet(GetExecState(L), db, L)
		},
		"kv_get_by_group": func(L *lua.LState) int {
			return kvGetByGroup(GetExecState(L), db, L)
		},
		"kv_remove": func(L *lua.LState) int {
			return kvRemove(GetExecState(L), db, L)
		},
		"kv_update": func(L *lua.LState) int {
			return kvUpdate(GetExecState(L), db, L)
		},
		"kv_upsert": func(L *lua.LState) int {
			return kvUpsert(GetExecState(L), db, L)
		},
	}
}

func kvQuery(es *executors.ExecState, db datahub.SpaceKVOps, L *lua.LState) int {
	query := &SpaceKVQuery{}
	err := luaplus.MapToStruct(L, L.CheckTable(1), query)
	if err != nil {
		return luaplus.PushError(L, err)
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
		datas, err = db.QueryWithValueSpaceKV(es.InstalledId, query.Cond, query.Offset, query.Limit)
	} else {
		datas, err = db.QuerySpaceKV(es.InstalledId, query.Cond, query.Offset, query.Limit)
	}
	if err != nil {
		return luaplus.PushError(L, err)
	}

	result := L.NewTable()
	for _, data := range datas {
		luaTable, err := luaplus.StructToTable(L, data)
		if err != nil {
			return luaplus.PushError(L, err)
		}
		result.Append(luaTable)
	}

	L.Push(result)
	return 1
}

func kvAdd(es *executors.ExecState, db datahub.SpaceKVOps, L *lua.LState) int {
	dataStruct := &dbmodels.SpaceKV{}

	luaTable, err := luaplus.StructToTable(L, dataStruct)
	if err != nil {
		return luaplus.PushError(L, err)
	}

	err = db.AddSpaceKV(es.InstalledId, dataStruct)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	L.Push(luaTable)
	return 1
}

func kvGet(es *executors.ExecState, db datahub.SpaceKVOps, L *lua.LState) int {
	group := L.CheckString(1)
	key := L.CheckString(2)
	data, err := db.GetSpaceKV(es.InstalledId, group, key)
	if err != nil {
		return luaplus.PushError(L, err)
	}

	luaTable, err := luaplus.StructToTable(L, data)
	if err != nil {
		return luaplus.PushError(L, err)
	}

	L.Push(luaTable)
	return 1
}

func kvGetByGroup(es *executors.ExecState, db datahub.SpaceKVOps, L *lua.LState) int {
	group := L.CheckString(1)
	offset := L.CheckInt(2)
	limit := L.CheckInt(3)
	datas, err := db.GetSpaceKVByGroup(es.InstalledId, group, offset, limit)
	if err != nil {
		return luaplus.PushError(L, err)
	}

	result := L.NewTable()
	for _, data := range datas {
		luaTable, err := luaplus.StructToTable(L, data)
		if err != nil {
			return luaplus.PushError(L, err)
		}
		result.Append(luaTable)
	}

	L.Push(result)
	return 1
}

func kvRemove(es *executors.ExecState, db datahub.SpaceKVOps, L *lua.LState) int {
	group := L.CheckString(1)
	key := L.CheckString(2)
	err := db.RemoveSpaceKV(es.InstalledId, group, key)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	L.Push(lua.LNil)
	return 1
}

func kvUpdate(es *executors.ExecState, db datahub.SpaceKVOps, L *lua.LState) int {
	group := L.CheckString(1)
	key := L.CheckString(2)
	data := L.CheckTable(3)
	dataMap := luaplus.TableToMap(L, data)

	err := db.UpdateSpaceKV(es.InstalledId, group, key, dataMap)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	L.Push(lua.LNil)
	return 1
}

func kvUpsert(es *executors.ExecState, db datahub.SpaceKVOps, L *lua.LState) int {
	group := L.CheckString(1)
	key := L.CheckString(2)
	data := L.CheckTable(3)
	dataMap := luaplus.TableToMap(L, data)
	err := db.UpsertSpaceKV(es.InstalledId, group, key, dataMap)
	if err != nil {
		return luaplus.PushError(L, err)
	}

	L.Push(lua.LNil)
	return 1
}
