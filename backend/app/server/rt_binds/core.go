package rtbinds

/*

core:
		"publish_event": func(L *lua.LState) int {
			return corePublishEvent(engine, GetExecState(L), L)
		},
		"file_token": func(L *lua.LState) int {
			return coreSignFsPresignedToken(sig, GetExecState(L), L)
		},
		"sign_advisery_token": func(L *lua.LState) int {
			return coreSignAdviseryToken(sig, GetExecState(L), L)
		},
		"parse_advisery_token": func(L *lua.LState) int {
			return coreParseAdviseryToken(sig, GetExecState(L), L)
		},
		"read_package_file": func(L *lua.LState) int {
			return readPackageFile(pops, GetExecState(L), L)
		},
		"list_files": func(L *lua.LState) int {
			coreHub := app.CoreHub().(*corehub.CoreHub)

			return coreListFiles(coreHub, GetExecState(L), L)
		},
		"decode_file_id": func(L *lua.LState) int {
			coreHub := app.CoreHub().(*corehub.CoreHub)
			return coreDecodeFileId(coreHub, L)
		},
		"encode_file_id": func(L *lua.LState) int {
			coreHub := app.CoreHub().(*corehub.CoreHub)
			return coreEncodeFileId(coreHub, L)
		},
		"db_vendor": func(L *lua.LState) int {
			L.Push(lua.LString(app.Database().Vender()))
			return 1
		},

		"get_env": func(L *lua.LState) int {
			return getEnv(app, L)
		},

*/
