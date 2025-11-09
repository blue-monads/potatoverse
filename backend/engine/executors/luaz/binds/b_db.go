package binds

import (
	"strconv"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/utils/luaplus"
	"github.com/blue-monads/turnix/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

func BindsDB(app xtypes.App, installId int64) func(L *lua.LState) int {
	return bindsDB(app, installId)
}

func bindsDB(app xtypes.App, installId int64) func(L *lua.LState) int {
	return func(L *lua.LState) int {

		db := app.Database().
			GetLowDBOps("P", strconv.FormatInt(installId, 10))

		table := L.NewTable()

		runDDL := func(L *lua.LState) int {
			ddl := L.CheckString(1)
			err := db.RunDDL(ddl)
			if err != nil {
				return pushError(L, err)
			}
			L.Push(lua.LNil)
			return 1
		}

		runQuery := func(L *lua.LState) int {
			query := L.CheckString(1)
			var args []any
			// Collect variadic arguments from index 2 onwards
			for i := 2; i <= L.GetTop(); i++ {
				arg := L.Get(i)
				args = append(args, luaplus.LuaTypeToGoType(L, arg))
			}
			results, err := db.RunQuery(query, args...)
			if err != nil {
				return pushError(L, err)
			}
			resultTable := L.NewTable()
			for _, row := range results {
				rowTable := luaplus.MapToTable(L, row)
				resultTable.Append(rowTable)
			}
			L.Push(resultTable)
			return 1
		}

		runQueryOne := func(L *lua.LState) int {
			query := L.CheckString(1)
			var args []any
			// Collect variadic arguments from index 2 onwards
			for i := 2; i <= L.GetTop(); i++ {
				arg := L.Get(i)
				args = append(args, luaplus.LuaTypeToGoType(L, arg))
			}
			result, err := db.RunQueryOne(query, args...)
			if err != nil {
				return pushError(L, err)
			}
			if result == nil {
				L.Push(lua.LNil)
				return 1
			}
			resultTable := luaplus.MapToTable(L, result)
			L.Push(resultTable)
			return 1
		}

		insert := func(L *lua.LState) int {
			tableName := L.CheckString(1)
			dataTable := L.CheckTable(2)
			dataMap := luaplus.TableToMap(L, dataTable)
			id, err := db.Insert(tableName, dataMap)
			if err != nil {
				return pushError(L, err)
			}
			L.Push(lua.LNumber(id))
			return 1
		}

		updateById := func(L *lua.LState) int {
			tableName := L.CheckString(1)
			id := int64(L.CheckNumber(2))
			dataTable := L.CheckTable(3)
			dataMap := luaplus.TableToMap(L, dataTable)
			err := db.UpdateById(tableName, id, dataMap)
			if err != nil {
				return pushError(L, err)
			}
			L.Push(lua.LNil)
			return 1
		}

		deleteById := func(L *lua.LState) int {
			tableName := L.CheckString(1)
			id := int64(L.CheckNumber(2))
			err := db.DeleteById(tableName, id)
			if err != nil {
				return pushError(L, err)
			}
			L.Push(lua.LNil)
			return 1
		}

		findById := func(L *lua.LState) int {
			tableName := L.CheckString(1)
			id := int64(L.CheckNumber(2))
			result, err := db.FindById(tableName, id)
			if err != nil {
				return pushError(L, err)
			}
			if result == nil {
				L.Push(lua.LNil)
				return 1
			}
			resultTable := luaplus.MapToTable(L, result)
			L.Push(resultTable)
			return 1
		}

		updateByCond := func(L *lua.LState) int {
			tableName := L.CheckString(1)
			condTable := L.CheckTable(2)
			dataTable := L.CheckTable(3)
			condMap := luaplus.TableToMapAny(L, condTable)
			dataMap := luaplus.TableToMap(L, dataTable)
			err := db.UpdateByCond(tableName, condMap, dataMap)
			if err != nil {
				return pushError(L, err)
			}
			L.Push(lua.LNil)
			return 1
		}

		deleteByCond := func(L *lua.LState) int {
			tableName := L.CheckString(1)
			condTable := L.CheckTable(2)
			condMap := luaplus.TableToMapAny(L, condTable)
			err := db.DeleteByCond(tableName, condMap)
			if err != nil {
				return pushError(L, err)
			}
			L.Push(lua.LNil)
			return 1
		}

		findAllByCond := func(L *lua.LState) int {
			tableName := L.CheckString(1)
			condTable := L.CheckTable(2)
			condMap := luaplus.TableToMapAny(L, condTable)
			results, err := db.FindAllByCond(tableName, condMap)
			if err != nil {
				return pushError(L, err)
			}
			resultTable := L.NewTable()
			for _, row := range results {
				rowTable := luaplus.MapToTable(L, row)
				resultTable.Append(rowTable)
			}
			L.Push(resultTable)
			return 1
		}

		findOneByCond := func(L *lua.LState) int {
			tableName := L.CheckString(1)
			condTable := L.CheckTable(2)
			condMap := luaplus.TableToMapAny(L, condTable)
			result, err := db.FindOneByCond(tableName, condMap)
			if err != nil {
				return pushError(L, err)
			}
			if result == nil {
				L.Push(lua.LNil)
				return 1
			}
			resultTable := luaplus.MapToTable(L, result)
			L.Push(resultTable)
			return 1
		}

		findAllByQuery := func(L *lua.LState) int {
			queryTable := L.CheckTable(1)
			query := &datahub.FindQuery{}
			err := luaplus.MapToStruct(L, queryTable, query)
			if err != nil {
				return pushError(L, err)
			}
			results, err := db.FindAllByQuery(query)
			if err != nil {
				return pushError(L, err)
			}
			resultTable := L.NewTable()
			for _, row := range results {
				rowTable := luaplus.MapToTable(L, row)
				resultTable.Append(rowTable)
			}
			L.Push(resultTable)
			return 1
		}

		findByJoin := func(L *lua.LState) int {
			queryTable := L.CheckTable(1)
			query := &datahub.FindByJoin{}
			err := luaplus.MapToStruct(L, queryTable, query)
			if err != nil {
				return pushError(L, err)
			}
			results, err := db.FindByJoin(query)
			if err != nil {
				return pushError(L, err)
			}
			resultTable := L.NewTable()
			for _, row := range results {
				rowTable := luaplus.MapToTable(L, row)
				resultTable.Append(rowTable)
			}
			L.Push(resultTable)
			return 1
		}

		listTables := func(L *lua.LState) int {
			tables, err := db.ListTables()
			if err != nil {
				return pushError(L, err)
			}
			resultTable := L.NewTable()
			for _, tableName := range tables {
				resultTable.Append(lua.LString(tableName))
			}
			L.Push(resultTable)
			return 1
		}

		listTableColumns := func(L *lua.LState) int {
			tableName := L.CheckString(1)
			columns, err := db.ListTableColumns(tableName)
			if err != nil {
				return pushError(L, err)
			}
			resultTable := L.NewTable()
			for _, column := range columns {
				columnTable := luaplus.MapToTable(L, column)
				resultTable.Append(columnTable)
			}
			L.Push(resultTable)
			return 1
		}

		L.SetFuncs(table, map[string]lua.LGFunction{
			"run_ddl":           runDDL,
			"run_query":         runQuery,
			"run_query_one":     runQueryOne,
			"insert":            insert,
			"update_by_id":      updateById,
			"delete_by_id":      deleteById,
			"find_by_id":        findById,
			"update_by_cond":    updateByCond,
			"delete_by_cond":    deleteByCond,
			"find_all_by_cond":  findAllByCond,
			"find_one_by_cond":  findOneByCond,
			"find_all_by_query": findAllByQuery,
			"find_by_join":      findByJoin,
			"list_tables":       listTables,
			"list_columns":      listTableColumns,
		})

		L.Push(table)
		return 1
	}
}
