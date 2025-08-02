package blua

import (
	"github.com/blue-monads/turnix/backend/engine/luaz/binds"
	lua "github.com/yuin/gopher-lua"
)

// Gin context methods exposed to Lua
func (b *Blua) reqAbort(L *lua.LState) int {
	b.handle.activeRequest.Abort()
	return 0
}

func (b *Blua) reqAbortWithStatus(L *lua.LState) int {
	code := L.CheckInt(1)
	b.handle.activeRequest.AbortWithStatus(code)
	return 0
}

func (b *Blua) reqAbortWithStatusJSON(L *lua.LState) int {
	code := L.CheckInt(1)
	jsonTbl := L.CheckTable(2)
	jsonObj := binds.TableToMap(L, jsonTbl)
	b.handle.activeRequest.AbortWithStatusJSON(code, jsonObj)
	return 0
}

func (b *Blua) reqAddParam(L *lua.LState) int {
	key := L.CheckString(1)
	value := L.CheckString(2)
	b.handle.activeRequest.AddParam(key, value)
	return 0
}

func (b *Blua) reqClientIP(L *lua.LState) int {
	L.Push(lua.LString(b.handle.activeRequest.ClientIP()))
	return 1
}

func (b *Blua) reqContentType(L *lua.LState) int {
	L.Push(lua.LString(b.handle.activeRequest.ContentType()))
	return 1
}

func (b *Blua) reqCookie(L *lua.LState) int {
	name := L.CheckString(1)
	value, err := b.handle.activeRequest.Cookie(name)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(value))
	return 1
}

func (b *Blua) reqData(L *lua.LState) int {
	code := L.CheckInt(1)
	contentType := L.CheckString(2)
	data := []byte(L.CheckString(3))
	b.handle.activeRequest.Data(code, contentType, data)
	return 0
}

func (b *Blua) reqDefaultQuery(L *lua.LState) int {
	key := L.CheckString(1)
	defaultValue := L.CheckString(2)
	L.Push(lua.LString(b.handle.activeRequest.DefaultQuery(key, defaultValue)))
	return 1
}

func (b *Blua) reqDefaultPostForm(L *lua.LState) int {
	key := L.CheckString(1)
	defaultValue := L.CheckString(2)
	L.Push(lua.LString(b.handle.activeRequest.DefaultPostForm(key, defaultValue)))
	return 1
}

func (b *Blua) reqFullPath(L *lua.LState) int {
	L.Push(lua.LString(b.handle.activeRequest.FullPath()))
	return 1
}

func (b *Blua) reqGetHeader(L *lua.LState) int {
	key := L.CheckString(1)
	L.Push(lua.LString(b.handle.activeRequest.GetHeader(key)))
	return 1
}

func (b *Blua) reqGetQuery(L *lua.LState) int {
	key := L.CheckString(1)
	value, exists := b.handle.activeRequest.GetQuery(key)
	L.Push(lua.LString(value))
	L.Push(lua.LBool(exists))
	return 2
}

func (b *Blua) reqGetPostForm(L *lua.LState) int {
	key := L.CheckString(1)
	value, exists := b.handle.activeRequest.GetPostForm(key)
	L.Push(lua.LString(value))
	L.Push(lua.LBool(exists))
	return 2
}

func (b *Blua) reqParam(L *lua.LState) int {
	key := L.CheckString(1)
	L.Push(lua.LString(b.handle.activeRequest.Param(key)))
	return 1
}

func (b *Blua) reqRedirect(L *lua.LState) int {
	code := L.CheckInt(1)
	location := L.CheckString(2)
	b.handle.activeRequest.Redirect(code, location)
	return 0
}

func (b *Blua) reqRemoteIP(L *lua.LState) int {
	L.Push(lua.LString(b.handle.activeRequest.RemoteIP()))
	return 1
}

func (b *Blua) reqJSON(L *lua.LState) int {
	code := L.CheckInt(1)
	jsonTbl := L.CheckTable(2)
	jsonObj := binds.TableToMap(L, jsonTbl)
	b.handle.activeRequest.JSON(code, jsonObj)
	return 0
}

func (b *Blua) reqHTML(L *lua.LState) int {
	code := L.CheckInt(1)
	name := L.CheckString(2)
	dataTbl := L.CheckTable(3)
	dataObj := binds.TableToMap(L, dataTbl)
	b.handle.activeRequest.HTML(code, name, dataObj)
	return 0
}

func (b *Blua) reqString(L *lua.LState) int {
	code := L.CheckInt(1)
	format := L.CheckString(2)

	// Get values for formatting
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

	b.handle.activeRequest.String(code, format, values...)
	return 0
}

func (b *Blua) reqSetCookie(L *lua.LState) int {
	name := L.CheckString(1)
	value := L.CheckString(2)
	maxAge := L.CheckInt(3)
	path := L.CheckString(4)
	domain := L.CheckString(5)
	secure := L.CheckBool(6)
	httpOnly := L.CheckBool(7)
	b.handle.activeRequest.SetCookie(name, value, maxAge, path, domain, secure, httpOnly)
	return 0
}

func (b *Blua) reqStatus(L *lua.LState) int {
	code := L.CheckInt(1)
	b.handle.activeRequest.Status(code)
	return 0
}

func (b *Blua) reqHeader(L *lua.LState) int {
	key := L.CheckString(1)
	value := L.CheckString(2)
	b.handle.activeRequest.Header(key, value)
	return 0
}

// Gin context binding methods
func (b *Blua) reqBindJSON(L *lua.LState) int {
	var obj map[string]any
	err := b.handle.activeRequest.BindJSON(&obj)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	result := binds.MapToTable(L, obj)
	L.Push(result)
	return 1
}

func (b *Blua) reqBindHeader(L *lua.LState) int {
	var obj map[string]any
	err := b.handle.activeRequest.BindHeader(&obj)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	result := binds.MapToTable(L, obj)
	L.Push(result)
	return 1
}

func (b *Blua) reqBindQuery(L *lua.LState) int {
	var obj map[string]any
	err := b.handle.activeRequest.BindQuery(&obj)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	result := binds.MapToTable(L, obj)
	L.Push(result)
	return 1
}

func (b *Blua) reqGetRawData(L *lua.LState) int {
	data, err := b.handle.activeRequest.GetRawData()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(string(data)))
	return 1
}

func (b *Blua) reqFormFile(L *lua.LState) int {
	name := L.CheckString(1)
	file, err := b.handle.activeRequest.FormFile(name)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Create a table to represent the file header
	fileTable := L.NewTable()
	L.SetField(fileTable, "filename", lua.LString(file.Filename))
	L.SetField(fileTable, "size", lua.LNumber(file.Size))

	// fixme
	// The actual file content needs to be read if needed
	// This is simplified since we can't easily pass the multipart.File to Lua
	L.Push(fileTable)
	return 1
}

func (b *Blua) reqGetQueryMap(L *lua.LState) int {
	key := L.CheckString(1)
	values, exists := b.handle.activeRequest.GetQueryMap(key)
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

func (b *Blua) reqGetQueryArray(L *lua.LState) int {
	key := L.CheckString(1)
	values, exists := b.handle.activeRequest.GetQueryArray(key)
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

func (b *Blua) reqGetPostFormMap(L *lua.LState) int {
	key := L.CheckString(1)
	values, exists := b.handle.activeRequest.GetPostFormMap(key)
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

func (b *Blua) reqGetPostFormArray(L *lua.LState) int {
	key := L.CheckString(1)
	values, exists := b.handle.activeRequest.GetPostFormArray(key)
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

func (b *Blua) reqSSEvent(L *lua.LState) int {
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
		msgValue = binds.TableToMap(L, message.(*lua.LTable))
	default:
		msgValue = message.String()
	}

	b.handle.activeRequest.SSEvent(name, msgValue)
	return 0
}

// Helper function to convert Go maps to Lua tables
