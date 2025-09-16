package binds

import (
	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
)

func HttpModule(L *lua.LState, ctx *gin.Context) *lua.LTable {

	mod := L.NewTable()

	reqAbort := func(L *lua.LState) int {
		ctx.Abort()
		return 0
	}

	reqAbortWithStatus := func(L *lua.LState) int {
		code := L.CheckInt(1)
		ctx.AbortWithStatus(code)
		return 0
	}

	reqAbortWithStatusJSON := func(L *lua.LState) int {
		code := L.CheckInt(1)
		jsonTbl := L.CheckTable(2)
		jsonObj := TableToMap(L, jsonTbl)
		ctx.AbortWithStatusJSON(code, jsonObj)
		return 0
	}

	reqAddParam := func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.CheckString(2)
		ctx.AddParam(key, value)
		return 0
	}

	reqClientIP := func(L *lua.LState) int {
		L.Push(lua.LString(ctx.ClientIP()))
		return 1
	}

	reqContentType := func(L *lua.LState) int {
		L.Push(lua.LString(ctx.ContentType()))
		return 1
	}

	reqCookie := func(L *lua.LState) int {
		name := L.CheckString(1)
		value, err := ctx.Cookie(name)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(value))
		return 1
	}

	reqData := func(L *lua.LState) int {
		code := L.CheckInt(1)
		contentType := L.CheckString(2)
		data := []byte(L.CheckString(3))
		ctx.Data(code, contentType, data)
		return 0
	}

	reqDefaultQuery := func(L *lua.LState) int {
		key := L.CheckString(1)
		defaultValue := L.CheckString(2)
		L.Push(lua.LString(ctx.DefaultQuery(key, defaultValue)))
		return 1
	}

	reqDefaultPostForm := func(L *lua.LState) int {
		key := L.CheckString(1)
		defaultValue := L.CheckString(2)
		L.Push(lua.LString(ctx.DefaultPostForm(key, defaultValue)))
		return 1
	}

	reqFullPath := func(L *lua.LState) int {
		L.Push(lua.LString(ctx.FullPath()))
		return 1
	}

	reqGetHeader := func(L *lua.LState) int {
		key := L.CheckString(1)
		L.Push(lua.LString(ctx.GetHeader(key)))
		return 1
	}

	reqGetQuery := func(L *lua.LState) int {
		key := L.CheckString(1)
		value, exists := ctx.GetQuery(key)
		if !exists {
			L.Push(lua.LNil)
			L.Push(lua.LBool(false))
			return 2
		}
		L.Push(lua.LString(value))
		L.Push(lua.LBool(true))
		return 2
	}

	reqGetPostForm := func(L *lua.LState) int {
		key := L.CheckString(1)
		value, exists := ctx.GetPostForm(key)
		if !exists {
			L.Push(lua.LNil)
			L.Push(lua.LBool(false))
			return 2
		}
		L.Push(lua.LString(value))
		L.Push(lua.LBool(true))
		return 2
	}

	reqParam := func(L *lua.LState) int {
		key := L.CheckString(1)
		L.Push(lua.LString(ctx.Param(key)))
		return 1
	}

	reqRedirect := func(L *lua.LState) int {
		code := L.CheckInt(1)
		location := L.CheckString(2)
		ctx.Redirect(code, location)
		return 0
	}

	reqRemoteIP := func(L *lua.LState) int {
		L.Push(lua.LString(ctx.ClientIP()))
		return 1
	}

	reqJSON := func(L *lua.LState) int {
		code := L.CheckInt(1)
		jsonTbl := L.CheckTable(2)
		jsonObj := TableToMap(L, jsonTbl)
		ctx.JSON(code, jsonObj)
		return 0
	}

	reqHTML := func(L *lua.LState) int {
		code := L.CheckInt(1)
		name := L.CheckString(2)
		dataTbl := L.CheckTable(3)
		dataObj := TableToMap(L, dataTbl)
		ctx.HTML(code, name, dataObj)
		return 0
	}

	reqString := func(L *lua.LState) int {
		code := L.CheckInt(1)
		format := L.CheckString(2)
		n := L.GetTop()
		values := make([]any, 0, n-2)
		for i := 3; i <= n; i++ {
			val := L.Get(i)
			switch val.Type() {
			case lua.LTString:
				values = append(values, val.String())
			case lua.LTNumber:
				values = append(values, float64(val.(lua.LNumber)))
			case lua.LTBool:
				values = append(values, bool(val.(lua.LBool)))
			default:
				values = append(values, val.String())
			}
		}
		ctx.String(code, format, values...)
		return 0
	}

	reqSetCookie := func(L *lua.LState) int {
		name := L.CheckString(1)
		value := L.CheckString(2)
		maxAge := L.CheckInt(3)
		path := L.CheckString(4)
		domain := L.CheckString(5)
		secure := L.CheckBool(6)
		httpOnly := L.CheckBool(7)
		ctx.SetCookie(name, value, maxAge, path, domain, secure, httpOnly)
		return 0
	}

	reqStatus := func(L *lua.LState) int {
		code := L.CheckInt(1)
		ctx.Status(code)
		return 0
	}

	reqHeader := func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.CheckString(2)
		ctx.Header(key, value)
		return 0
	}

	reqBindJSON := func(L *lua.LState) int {
		var obj map[string]any
		err := ctx.BindJSON(&obj)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		result := MapToTable(L, obj)
		L.Push(result)
		return 1
	}

	reqBindHeader := func(L *lua.LState) int {
		var obj map[string]any
		err := ctx.BindHeader(&obj)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		result := MapToTable(L, obj)
		L.Push(result)
		return 1
	}

	reqBindQuery := func(L *lua.LState) int {
		var obj map[string]any
		err := ctx.BindQuery(&obj)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		result := MapToTable(L, obj)
		L.Push(result)
		return 1

	}

	reqGetRawData := func(L *lua.LState) int {
		data, err := ctx.GetRawData()
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(string(data)))
		return 1
	}

	reqFormFile := func(L *lua.LState) int {
		name := L.CheckString(1)
		file, err := ctx.FormFile(name)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		fileTable := L.NewTable()
		L.SetField(fileTable, "filename", lua.LString(file.Filename))
		L.SetField(fileTable, "size", lua.LNumber(file.Size))
		L.Push(fileTable)
		return 1
	}

	reqGetQueryMap := func(L *lua.LState) int {
		key := L.CheckString(1)
		values, exists := ctx.GetQueryMap(key)
		if !exists {
			L.Push(lua.LNil)
			L.Push(lua.LBool(false))
			return 2
		}
		table := L.NewTable()
		for k, v := range values {
			L.SetField(table, k, lua.LString(v))
		}
		L.Push(table)
		L.Push(lua.LBool(true))
		return 2
	}

	reqGetQueryArray := func(L *lua.LState) int {
		key := L.CheckString(1)
		values, exists := ctx.GetQueryArray(key)
		if !exists {
			L.Push(lua.LNil)
			L.Push(lua.LBool(false))
			return 2
		}
		table := L.NewTable()
		for i, v := range values {
			L.RawSetInt(table, i+1, lua.LString(v))
		}
		L.Push(table)
		L.Push(lua.LBool(true))
		return 2
	}

	reqGetPostFormMap := func(L *lua.LState) int {
		key := L.CheckString(1)
		values, exists := ctx.GetPostFormMap(key)
		if !exists {
			L.Push(lua.LNil)
			L.Push(lua.LBool(false))
			return 2
		}
		table := L.NewTable()
		for k, v := range values {
			L.SetField(table, k, lua.LString(v))
		}
		L.Push(table)
		L.Push(lua.LBool(true))
		return 2
	}

	reqGetPostFormArray := func(L *lua.LState) int {
		key := L.CheckString(1)
		values, exists := ctx.GetPostFormArray(key)
		if !exists {
			L.Push(lua.LNil)
			L.Push(lua.LBool(false))
			return 2
		}
		table := L.NewTable()
		for i, v := range values {
			L.RawSetInt(table, i+1, lua.LString(v))
		}
		L.Push(table)
		L.Push(lua.LBool(true))
		return 2
	}

	reqSSEvent := func(L *lua.LState) int {
		name := L.CheckString(1)
		message := L.CheckAny(2)

		var msgValue any

		switch message.Type() {
		case lua.LTString:
			msgValue = message.String()
		case lua.LTNumber:
			msgValue = float64(message.(lua.LNumber))
		case lua.LTBool:
			msgValue = bool(message.(lua.LBool))
		case lua.LTTable:
			msgValue = TableToMap(L, message.(*lua.LTable))
		default:
			msgValue = message.String()
		}

		ctx.SSEvent(name, msgValue)
		return 0
	}

	L.SetFuncs(mod, map[string]lua.LGFunction{
		"abort":               reqAbort,
		"abortWithStatus":     reqAbortWithStatus,
		"abortWithStatusJSON": reqAbortWithStatusJSON,
		"addParam":            reqAddParam,
		"clientIP":            reqClientIP,
		"contentType":         reqContentType,
		"cookie":              reqCookie,
		"data":                reqData,
		"defaultQuery":        reqDefaultQuery,
		"defaultPostForm":     reqDefaultPostForm,
		"fullPath":            reqFullPath,
		"getHeader":           reqGetHeader,
		"getQuery":            reqGetQuery,
		"getPostForm":         reqGetPostForm,
		"param":               reqParam,
		"redirect":            reqRedirect,
		"remoteIP":            reqRemoteIP,
		"json":                reqJSON,
		"html":                reqHTML,
		"string":              reqString,
		"setCookie":           reqSetCookie,
		"status":              reqStatus,
		"header":              reqHeader,
		"bindJSON":            reqBindJSON,
		"bindHeader":          reqBindHeader,
		"bindQuery":           reqBindQuery,
		"getRawData":          reqGetRawData,
		"formFile":            reqFormFile,
		"getQueryMap":         reqGetQueryMap,
		"getQueryArray":       reqGetQueryArray,
		"getPostFormMap":      reqGetPostFormMap,
		"getPostFormArray":    reqGetPostFormArray,
		"sseEvent":            reqSSEvent,
	})

	return mod

}
