package binds

import (
	"github.com/blue-monads/turnix/backend/engine/executors"
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	"github.com/blue-monads/turnix/backend/utils/luaplus"
	lua "github.com/yuin/gopher-lua"
)

func pushError(L *lua.LState, err error) int {
	return luaplus.PushError(L, err)
}

func bindsKV(spaceId int64, db datahub.SpaceKVOps) func(L *lua.LState) int {

	return func(L *lua.LState) int {

		QuerySpaceKV := func(L *lua.LState) int {

			cond := L.CheckTable(1)

			condMap := luaplus.TableToMapAny(L, cond)

			datas, err := db.QuerySpaceKV(spaceId, condMap)
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
		AddSpaceKV := func(L *lua.LState) int {

			dataStruct := &dbmodels.SpaceKV{}

			luaTable, err := luaplus.StructToTable(L, dataStruct)
			if err != nil {
				return pushError(L, err)
			}

			err = db.AddSpaceKV(spaceId, dataStruct)
			if err != nil {
				return pushError(L, err)
			}
			L.Push(luaTable)
			return 1
		}

		GetSpaceKV := func(L *lua.LState) int {
			group := L.CheckString(1)
			key := L.CheckString(2)
			data, err := db.GetSpaceKV(spaceId, group, key)
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

		GetSpaceKVByGroup := func(L *lua.LState) int {
			group := L.CheckString(1)
			offset := L.CheckInt(2)
			limit := L.CheckInt(3)
			datas, err := db.GetSpaceKVByGroup(spaceId, group, offset, limit)
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

		RemoveSpaceKV := func(L *lua.LState) int {
			group := L.CheckString(1)
			key := L.CheckString(2)
			err := db.RemoveSpaceKV(spaceId, group, key)
			if err != nil {
				return pushError(L, err)
			}
			return 1
		}

		UpdateSpaceKV := func(L *lua.LState) int {
			group := L.CheckString(1)
			key := L.CheckString(2)
			data := L.CheckTable(3)
			dataMap := luaplus.TableToMap(L, data)

			err := db.UpdateSpaceKV(spaceId, group, key, dataMap)
			if err != nil {
				return pushError(L, err)
			}
			return 1
		}

		UpsertSpaceKV := func(L *lua.LState) int {
			group := L.CheckString(1)
			key := L.CheckString(2)
			data := L.CheckTable(3)
			dataMap := luaplus.TableToMap(L, data)
			err := db.UpsertSpaceKV(spaceId, group, key, dataMap)
			if err != nil {
				return pushError(L, err)
			}

			return 1
		}

		table := L.NewTable()
		L.SetFuncs(table, map[string]lua.LGFunction{
			"query":        QuerySpaceKV,
			"add":          AddSpaceKV,
			"get":          GetSpaceKV,
			"get_by_group": GetSpaceKVByGroup,
			"remove":       RemoveSpaceKV,
			"update":       UpdateSpaceKV,
			"upsert":       UpsertSpaceKV,
		})
		L.Push(table)
		return 1

	}

}

func BindsKV(spaceId int64, handle *executors.EHandle) func(L *lua.LState) int {
	return bindsKV(spaceId, handle.App.Database())
}
