package binds

import (
	"strconv"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/utils/luaplus"
	"github.com/blue-monads/turnix/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

// DB Module
func registerDBModuleType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaDBModuleTypeName)
	L.SetField(mt, "__index", L.NewFunction(dbModuleIndex))
}

func newDBModule(L *lua.LState, app xtypes.App, installId int64) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = &luaDBModule{
		app:       app,
		installId: installId,
		db:        app.Database().GetLowDBOps("P", strconv.FormatInt(installId, 10)),
	}
	L.SetMetatable(ud, L.GetTypeMetatable(luaDBModuleTypeName))
	return ud
}

func checkDBModule(L *lua.LState) *luaDBModule {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*luaDBModule); ok {
		return v
	}
	L.ArgError(1, luaDBModuleTypeName+" expected")
	return nil
}

func dbModuleIndex(L *lua.LState) int {
	mod := checkDBModule(L)
	method := L.CheckString(2)

	switch method {
	case "run_ddl":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return dbRunDDL(mod, L)
		}))
		return 1
	case "run_query":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return dbRunQuery(mod, L)
		}))
		return 1
	case "run_query_one":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return dbRunQueryOne(mod, L)
		}))
		return 1
	case "insert":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return dbInsert(mod, L)
		}))
		return 1
	case "update_by_id":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return dbUpdateById(mod, L)
		}))
		return 1
	case "delete_by_id":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return dbDeleteById(mod, L)
		}))
		return 1
	case "find_by_id":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return dbFindById(mod, L)
		}))
		return 1
	case "update_by_cond":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return dbUpdateByCond(mod, L)
		}))
		return 1
	case "delete_by_cond":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return dbDeleteByCond(mod, L)
		}))
		return 1
	case "find_all_by_cond":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return dbFindAllByCond(mod, L)
		}))
		return 1
	case "find_one_by_cond":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return dbFindOneByCond(mod, L)
		}))
		return 1
	case "find_all_by_query":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return dbFindAllByQuery(mod, L)
		}))
		return 1
	case "find_by_join":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return dbFindByJoin(mod, L)
		}))
		return 1
	case "list_tables":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return dbListTables(mod, L)
		}))
		return 1
	case "list_columns":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return dbListTableColumns(mod, L)
		}))
		return 1
	}

	return 0
}

func dbRunDDL(mod *luaDBModule, L *lua.LState) int {
	ddl := L.CheckString(1)
	err := mod.db.RunDDL(ddl)
	if err != nil {
		return pushError(L, err)
	}
	L.Push(lua.LNil)
	return 1
}

func dbRunQuery(mod *luaDBModule, L *lua.LState) int {
	query := L.CheckString(1)
	var args []any
	for i := 2; i <= L.GetTop(); i++ {
		arg := L.Get(i)
		args = append(args, luaplus.LuaTypeToGoType(L, arg))
	}
	results, err := mod.db.RunQuery(query, args...)
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

func dbRunQueryOne(mod *luaDBModule, L *lua.LState) int {
	query := L.CheckString(1)
	var args []any
	for i := 2; i <= L.GetTop(); i++ {
		arg := L.Get(i)
		args = append(args, luaplus.LuaTypeToGoType(L, arg))
	}
	result, err := mod.db.RunQueryOne(query, args...)
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

func dbInsert(mod *luaDBModule, L *lua.LState) int {
	tableName := L.CheckString(1)
	dataTable := L.CheckTable(2)
	dataMap := luaplus.TableToMap(L, dataTable)
	id, err := mod.db.Insert(tableName, dataMap)
	if err != nil {
		return pushError(L, err)
	}
	L.Push(lua.LNumber(id))
	return 1
}

func dbUpdateById(mod *luaDBModule, L *lua.LState) int {
	tableName := L.CheckString(1)
	id := int64(L.CheckNumber(2))
	dataTable := L.CheckTable(3)
	dataMap := luaplus.TableToMap(L, dataTable)
	err := mod.db.UpdateById(tableName, id, dataMap)
	if err != nil {
		return pushError(L, err)
	}
	L.Push(lua.LNil)
	return 1
}

func dbDeleteById(mod *luaDBModule, L *lua.LState) int {
	tableName := L.CheckString(1)
	id := int64(L.CheckNumber(2))
	err := mod.db.DeleteById(tableName, id)
	if err != nil {
		return pushError(L, err)
	}
	L.Push(lua.LNil)
	return 1
}

func dbFindById(mod *luaDBModule, L *lua.LState) int {
	tableName := L.CheckString(1)
	id := int64(L.CheckNumber(2))
	result, err := mod.db.FindById(tableName, id)
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

func dbUpdateByCond(mod *luaDBModule, L *lua.LState) int {
	tableName := L.CheckString(1)
	condTable := L.CheckTable(2)
	dataTable := L.CheckTable(3)
	condMap := luaplus.TableToMapAny(L, condTable)
	dataMap := luaplus.TableToMap(L, dataTable)
	err := mod.db.UpdateByCond(tableName, condMap, dataMap)
	if err != nil {
		return pushError(L, err)
	}
	L.Push(lua.LNil)
	return 1
}

func dbDeleteByCond(mod *luaDBModule, L *lua.LState) int {
	tableName := L.CheckString(1)
	condTable := L.CheckTable(2)
	condMap := luaplus.TableToMapAny(L, condTable)
	err := mod.db.DeleteByCond(tableName, condMap)
	if err != nil {
		return pushError(L, err)
	}
	L.Push(lua.LNil)
	return 1
}

func dbFindAllByCond(mod *luaDBModule, L *lua.LState) int {
	tableName := L.CheckString(1)
	condTable := L.CheckTable(2)
	condMap := luaplus.TableToMapAny(L, condTable)
	results, err := mod.db.FindAllByCond(tableName, condMap)
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

func dbFindOneByCond(mod *luaDBModule, L *lua.LState) int {
	tableName := L.CheckString(1)
	condTable := L.CheckTable(2)
	condMap := luaplus.TableToMapAny(L, condTable)
	result, err := mod.db.FindOneByCond(tableName, condMap)
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

func dbFindAllByQuery(mod *luaDBModule, L *lua.LState) int {
	queryTable := L.CheckTable(1)
	query := &datahub.FindQuery{}
	err := luaplus.MapToStruct(L, queryTable, query)
	if err != nil {
		return pushError(L, err)
	}
	results, err := mod.db.FindAllByQuery(query)
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

func dbFindByJoin(mod *luaDBModule, L *lua.LState) int {
	queryTable := L.CheckTable(1)
	query := &datahub.FindByJoin{}
	err := luaplus.MapToStruct(L, queryTable, query)
	if err != nil {
		return pushError(L, err)
	}
	results, err := mod.db.FindByJoin(query)
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

func dbListTables(mod *luaDBModule, L *lua.LState) int {
	tables, err := mod.db.ListTables()
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

func dbListTableColumns(mod *luaDBModule, L *lua.LState) int {
	tableName := L.CheckString(1)
	columns, err := mod.db.ListTableColumns(tableName)
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
