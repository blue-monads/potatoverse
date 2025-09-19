package binds

import (
	"reflect"

	"github.com/blue-monads/turnix/backend/engine/bhandle"
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/datahub/models"
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

			L.CheckUserData(1)

			dataStruct := &models.SpaceKV{}

			if err := toStructFromTableInner(L, L.CheckTable(1), reflect.ValueOf(dataStruct)); err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}

			err := db.AddSpaceKV(spaceId, dataStruct)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			return 1
		}

		table := L.NewTable()
		L.SetFuncs(table, map[string]lua.LGFunction{
			"QuerySpaceKV": QuerySpaceKV,
			"AddSpaceKV":   AddSpaceKV,
		})
		L.Push(table)
		return 1

	}

}

func BindsDB(handle *bhandle.Bhandle) func(L *lua.LState) int {
	return bindsDB(handle.SpaceId, handle.Database)
}
