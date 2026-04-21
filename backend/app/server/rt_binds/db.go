package rtbinds

/*





db:

		"vender": func(L *lua.LState) int {
			vender := db.Vender()
			L.Push(lua.LString(vender))
			return 1
		},
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


*/
