package binds2

import (
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/utils/luaplus"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

func DBBindable(app xtypes.App) map[string]lua.LGFunction {
	db := app.Database()

	getPackageDbOps := func(L *lua.LState) datahub.DBLowOps {
		es := GetExecState(L)
		installId := es.InstalledId
		return db.GetLowPackageDBOps(installId)
	}

	return map[string]lua.LGFunction{
		"run_ddl": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbRunDDL(dbOps, L)
		},
		"run_query": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbRunQuery(dbOps, L)
		},
		"run_query_one": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbRunQueryOne(dbOps, L)
		},
		"insert": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbInsert(dbOps, L)
		},
		"update_by_id": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbUpdateById(dbOps, L)
		},
		"delete_by_id": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbDeleteById(dbOps, L)
		},
		"find_by_id": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbFindById(dbOps, L)
		},
		"update_by_cond": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbUpdateByCond(dbOps, L)
		},
		"delete_by_cond": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbDeleteByCond(dbOps, L)
		},
		"find_all_by_cond": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbFindAllByCond(dbOps, L)
		},
		"find_one_by_cond": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbFindOneByCond(dbOps, L)
		},
		"find_all_by_query": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbFindAllByQuery(dbOps, L)
		},
		"find_by_join": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbFindByJoin(dbOps, L)
		},
		"list_tables": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbListTables(dbOps, L)
		},
		"list_columns": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbListTableColumns(dbOps, L)
		},
		"start_txn": func(L *lua.LState) int {
			dbOps := getPackageDbOps(L)
			return dbStartTxn(dbOps, L)
		},
	}

}

func dbStartTxn(dbOps datahub.DBLowOps, L *lua.LState) int {
	txn, err := dbOps.StartTxn()
	if err != nil {
		return luaplus.PushError(L, err)
	}

	txnTable := L.NewTable()

	L.SetFuncs(txnTable, map[string]lua.LGFunction{
		"run_ddl": func(L *lua.LState) int {
			return dbRunDDL(txn, L)
		},
		"run_query": func(L *lua.LState) int {
			return dbRunQuery(txn, L)
		},
		"run_query_one": func(L *lua.LState) int {
			return dbRunQueryOne(txn, L)
		},
		"insert": func(L *lua.LState) int {
			return dbInsert(txn, L)
		},
		"update_by_id": func(L *lua.LState) int {
			return dbUpdateById(txn, L)
		},
		"delete_by_id": func(L *lua.LState) int {
			return dbDeleteById(txn, L)
		},
		"find_by_id": func(L *lua.LState) int {
			return dbFindById(txn, L)
		},
		"update_by_cond": func(L *lua.LState) int {
			return dbUpdateByCond(txn, L)
		},
		"delete_by_cond": func(L *lua.LState) int {
			return dbDeleteByCond(txn, L)
		},
		"find_all_by_cond": func(L *lua.LState) int {
			return dbFindAllByCond(txn, L)
		},
		"find_one_by_cond": func(L *lua.LState) int {
			return dbFindOneByCond(txn, L)
		},
		"find_all_by_query": func(L *lua.LState) int {
			return dbFindAllByQuery(txn, L)
		},
		"find_by_join": func(L *lua.LState) int {
			return dbFindByJoin(txn, L)
		},
		"commit": func(L *lua.LState) int {
			return dbCommit(txn, L)
		},
		"rollback": func(L *lua.LState) int {
			return dbRollback(txn, L)
		},
	})
	L.Push(txnTable)

	return 1
}

func dbCommit(dbOps datahub.DBLowTxnOps, L *lua.LState) int {
	err := dbOps.Commit()
	if err != nil {
		return luaplus.PushError(L, err)
	}
	L.Push(lua.LNil)
	L.Push(lua.LNil)
	return 2
}

func dbRollback(dbOps datahub.DBLowTxnOps, L *lua.LState) int {
	err := dbOps.Rollback()
	if err != nil {
		return luaplus.PushError(L, err)
	}
	L.Push(lua.LNil)
	L.Push(lua.LNil)
	return 2
}

func dbRunDDL(dbOps datahub.DBLowCoreOps, L *lua.LState) int {
	ddl := L.CheckString(1)
	err := dbOps.RunDDL(ddl)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	L.Push(lua.LNil)
	L.Push(lua.LNil)
	return 2
}

func dbRunQuery(dbOps datahub.DBLowCoreOps, L *lua.LState) int {
	query := L.CheckString(1)
	var args []any
	for i := 2; i <= L.GetTop(); i++ {
		arg := L.Get(i)
		args = append(args, luaplus.LuaTypeToGoType(L, arg))
	}
	results, err := dbOps.RunQuery(query, args...)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	resultTable := L.NewTable()
	for _, row := range results {
		rowTable := luaplus.MapToTable(L, row)
		resultTable.Append(rowTable)
	}
	L.Push(resultTable)
	return 1
}

func dbRunQueryOne(dbOps datahub.DBLowCoreOps, L *lua.LState) int {
	query := L.CheckString(1)
	var args []any
	for i := 2; i <= L.GetTop(); i++ {
		arg := L.Get(i)
		args = append(args, luaplus.LuaTypeToGoType(L, arg))
	}
	result, err := dbOps.RunQueryOne(query, args...)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	if result == nil {
		L.Push(lua.LNil)
		return 1
	}
	resultTable := luaplus.MapToTable(L, result)
	L.Push(resultTable)
	return 1
}

func dbInsert(dbOps datahub.DBLowCoreOps, L *lua.LState) int {
	tableName := L.CheckString(1)
	dataTable := L.CheckTable(2)
	dataMap := luaplus.TableToMap(L, dataTable)
	id, err := dbOps.Insert(tableName, dataMap)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	L.Push(lua.LNumber(id))
	return 1
}

func dbUpdateById(dbOps datahub.DBLowCoreOps, L *lua.LState) int {
	tableName := L.CheckString(1)
	id := int64(L.CheckNumber(2))
	dataTable := L.CheckTable(3)
	dataMap := luaplus.TableToMap(L, dataTable)
	err := dbOps.UpdateById(tableName, id, dataMap)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	L.Push(lua.LNil)
	return 1
}

func dbDeleteById(dbOps datahub.DBLowCoreOps, L *lua.LState) int {
	tableName := L.CheckString(1)
	id := int64(L.CheckNumber(2))
	err := dbOps.DeleteById(tableName, id)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	L.Push(lua.LNil)
	return 1
}

func dbFindById(dbOps datahub.DBLowCoreOps, L *lua.LState) int {
	tableName := L.CheckString(1)
	id := int64(L.CheckNumber(2))
	result, err := dbOps.FindById(tableName, id)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	if result == nil {
		L.Push(lua.LNil)
		return 1
	}
	resultTable := luaplus.MapToTable(L, result)
	L.Push(resultTable)
	return 1
}

func dbUpdateByCond(dbOps datahub.DBLowCoreOps, L *lua.LState) int {
	tableName := L.CheckString(1)
	condTable := L.CheckTable(2)
	dataTable := L.CheckTable(3)
	condMap := luaplus.TableToMapAny(L, condTable)
	dataMap := luaplus.TableToMap(L, dataTable)
	err := dbOps.UpdateByCond(tableName, condMap, dataMap)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	L.Push(lua.LNil)
	return 1
}

func dbDeleteByCond(dbOps datahub.DBLowCoreOps, L *lua.LState) int {
	tableName := L.CheckString(1)
	condTable := L.CheckTable(2)
	condMap := luaplus.TableToMapAny(L, condTable)
	err := dbOps.DeleteByCond(tableName, condMap)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	L.Push(lua.LNil)
	return 1
}

func dbFindAllByCond(dbOps datahub.DBLowCoreOps, L *lua.LState) int {
	tableName := L.CheckString(1)
	condTable := L.CheckTable(2)
	condMap := luaplus.TableToMapAny(L, condTable)
	results, err := dbOps.FindAllByCond(tableName, condMap)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	resultTable := L.NewTable()
	for _, row := range results {
		rowTable := luaplus.MapToTable(L, row)
		resultTable.Append(rowTable)
	}
	L.Push(resultTable)
	return 1
}

func dbFindOneByCond(dbOps datahub.DBLowCoreOps, L *lua.LState) int {
	tableName := L.CheckString(1)
	condTable := L.CheckTable(2)
	condMap := luaplus.TableToMapAny(L, condTable)
	result, err := dbOps.FindOneByCond(tableName, condMap)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	if result == nil {
		L.Push(lua.LNil)
		return 1
	}
	resultTable := luaplus.MapToTable(L, result)
	L.Push(resultTable)
	return 1
}

func dbFindAllByQuery(dbOps datahub.DBLowCoreOps, L *lua.LState) int {
	queryTable := L.CheckTable(1)
	query := &datahub.FindQuery{}
	err := luaplus.MapToStruct(L, queryTable, query)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	results, err := dbOps.FindAllByQuery(query)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	resultTable := L.NewTable()
	for _, row := range results {
		rowTable := luaplus.MapToTable(L, row)
		resultTable.Append(rowTable)
	}
	L.Push(resultTable)
	return 1
}

func dbFindByJoin(dbOps datahub.DBLowCoreOps, L *lua.LState) int {
	queryTable := L.CheckTable(1)
	query := &datahub.FindByJoin{}
	err := luaplus.MapToStruct(L, queryTable, query)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	results, err := dbOps.FindByJoin(query)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	resultTable := L.NewTable()
	for _, row := range results {
		rowTable := luaplus.MapToTable(L, row)
		resultTable.Append(rowTable)
	}
	L.Push(resultTable)
	return 1
}

func dbListTables(dbOps datahub.DBLowOps, L *lua.LState) int {
	tables, err := dbOps.ListTables()
	if err != nil {
		return luaplus.PushError(L, err)
	}
	resultTable := L.NewTable()
	for _, tableName := range tables {
		resultTable.Append(lua.LString(tableName))
	}
	L.Push(resultTable)
	return 1
}

func dbListTableColumns(dbOps datahub.DBLowOps, L *lua.LState) int {
	tableName := L.CheckString(1)
	columns, err := dbOps.ListTableColumns(tableName)
	if err != nil {
		return luaplus.PushError(L, err)
	}
	resultTable := L.NewTable()
	for _, column := range columns {
		columnTable := luaplus.MapToTable(L, column)
		resultTable.Append(columnTable)
	}
	L.Push(resultTable)
	return 1
}
