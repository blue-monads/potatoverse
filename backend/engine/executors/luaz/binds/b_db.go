package binds

import (
	"reflect"

	"github.com/blue-monads/turnix/backend/engine/executors"
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/dbmodels"
	lua "github.com/yuin/gopher-lua"
)

func bindsDB(spaceId int64, db datahub.SpaceKVOps) func(L *lua.LState) int {

	return func(L *lua.LState) int {

		QuerySpaceKV := func(L *lua.LState) int {

			cond := L.CheckTable(1)

			condMap := TableToMapAny(L, cond)

			datas, err := db.QuerySpaceKV(spaceId, condMap)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			result := L.NewTable()
			for _, data := range datas {
				result.Append(ToTableFromStruct(L, reflect.ValueOf(data)))
			}

			L.Push(result)

			return 1
		}
		AddSpaceKV := func(L *lua.LState) int {

			dataStruct := dbmodels.SpaceKV{}

			if err := toStructFromTableInner(L, L.CheckTable(1), reflect.ValueOf(&dataStruct)); err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			err := db.AddSpaceKV(spaceId, &dataStruct)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			return 1
		}

		GetSpaceKV := func(L *lua.LState) int {
			group := L.CheckString(1)
			key := L.CheckString(2)
			data, err := db.GetSpaceKV(spaceId, group, key)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			L.Push(ToTableFromStruct(L, reflect.ValueOf(data)))
			return 1
		}

		GetSpaceKVByGroup := func(L *lua.LState) int {
			group := L.CheckString(1)
			offset := L.CheckInt(2)
			limit := L.CheckInt(3)
			datas, err := db.GetSpaceKVByGroup(spaceId, group, offset, limit)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			result := L.NewTable()
			for _, data := range datas {
				result.Append(ToTableFromStruct(L, reflect.ValueOf(data)))
			}

			L.Push(result)
			return 1
		}

		RemoveSpaceKV := func(L *lua.LState) int {
			group := L.CheckString(1)
			key := L.CheckString(2)
			err := db.RemoveSpaceKV(spaceId, group, key)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			return 1
		}

		UpdateSpaceKV := func(L *lua.LState) int {
			group := L.CheckString(1)
			key := L.CheckString(2)
			data := L.CheckTable(3)
			dataMap := TableToMap(L, data)

			err := db.UpdateSpaceKV(spaceId, group, key, dataMap)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			return 1
		}

		UpsertSpaceKV := func(L *lua.LState) int {
			group := L.CheckString(1)
			key := L.CheckString(2)
			data := L.CheckTable(3)
			dataMap := TableToMap(L, data)
			err := db.UpsertSpaceKV(spaceId, group, key, dataMap)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
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

func BindsDB(handle *executors.EHandle) func(L *lua.LState) int {
	return bindsDB(handle.SpaceId, handle.App.Database())
}
