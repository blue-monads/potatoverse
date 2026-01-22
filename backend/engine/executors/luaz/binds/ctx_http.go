package binds

import (
	"errors"

	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/utils/luaplus"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
)

const (
	luaHttpRequestContextTypeName = "http.request"
)

type luaHttpRequestContext struct {
	app        xtypes.App
	spaceId    int64
	ctx        *gin.Context
	sig        *signer.Signer
	spaceClaim *signer.SpaceClaim
}

// HTTP Request Context
func registerHttpRequestContextType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaHttpRequestContextTypeName)
	L.SetField(mt, "__index", L.NewFunction(httpRequestContextIndex))
}

func NewHttpRequestContext(L *lua.LState, app xtypes.App, spaceId int64, ctx *gin.Context) *lua.LUserData {
	registerHttpRequestContextType(L)
	mt := L.GetTypeMetatable(luaHttpRequestContextTypeName)

	ud := L.NewUserData()
	ud.Value = &luaHttpRequestContext{
		app:        app,
		spaceId:    spaceId,
		ctx:        ctx,
		sig:        app.Signer(),
		spaceClaim: nil,
	}
	L.SetMetatable(ud, mt)
	return ud
}

func checkHttpRequestContext(L *lua.LState) *luaHttpRequestContext {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*luaHttpRequestContext); ok {
		return v
	}
	L.ArgError(1, luaHttpRequestContextTypeName+" expected")
	return nil
}

func httpRequestContextIndex(L *lua.LState) int {
	reqCtx := checkHttpRequestContext(L)
	method := L.CheckString(2)

	switch method {
	case "abort":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqAbort(reqCtx, L)
		}))
		return 1
	case "abort_with_status":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqAbortWithStatus(reqCtx, L)
		}))
		return 1
	case "abort_with_status_json":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqAbortWithStatusJSON(reqCtx, L)
		}))
		return 1
	case "client_ip":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqClientIP(reqCtx, L)
		}))
		return 1
	case "content_type":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqContentType(reqCtx, L)
		}))
		return 1
	case "cookie":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqCookie(reqCtx, L)
		}))
		return 1
	case "data":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqData(reqCtx, L)
		}))
		return 1
	case "get_claim":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqGetClaim(reqCtx, L)
		}))
		return 1
	case "get_user_id":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqGetUserId(reqCtx, L)
		}))
		return 1
	case "default_query":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqDefaultQuery(reqCtx, L)
		}))
		return 1
	case "default_post_form":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqDefaultPostForm(reqCtx, L)
		}))
		return 1
	case "full_path":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqFullPath(reqCtx, L)
		}))
		return 1
	case "get_header":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqGetHeader(reqCtx, L)
		}))
		return 1
	case "get_query":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqGetQuery(reqCtx, L)
		}))
		return 1
	case "get_post_form":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqGetPostForm(reqCtx, L)
		}))
		return 1
	case "param":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqParam(reqCtx, L)
		}))
		return 1
	case "redirect":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqRedirect(reqCtx, L)
		}))
		return 1
	case "remote_ip":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqRemoteIP(reqCtx, L)
		}))
		return 1
	case "json":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqJSON(reqCtx, L)
		}))
		return 1
	case "json_array":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqJSONArray(reqCtx, L)
		}))
		return 1
	case "html":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqHTML(reqCtx, L)
		}))
		return 1
	case "string":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqString(reqCtx, L)
		}))
		return 1
	case "set_cookie":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqSetCookie(reqCtx, L)
		}))
		return 1
	case "status":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqStatus(reqCtx, L)
		}))
		return 1
	case "header":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqHeader(reqCtx, L)
		}))
		return 1
	case "bind_json":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqBindJSON(reqCtx, L)
		}))
		return 1
	case "bind_header":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqBindHeader(reqCtx, L)
		}))
		return 1
	case "bind_query":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqBindQuery(reqCtx, L)
		}))
		return 1
	case "get_raw_data":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqGetRawData(reqCtx, L)
		}))
		return 1
	case "form_file":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqFormFile(reqCtx, L)
		}))
		return 1
	case "get_query_map":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqGetQueryMap(reqCtx, L)
		}))
		return 1
	case "get_query_array":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqGetQueryArray(reqCtx, L)
		}))
		return 1
	case "get_post_form_map":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqGetPostFormMap(reqCtx, L)
		}))
		return 1
	case "get_post_form_array":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqGetPostFormArray(reqCtx, L)
		}))
		return 1
	case "sse_event":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqSSEvent(reqCtx, L)
		}))
		return 1
	case "state_keys":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqStateKeys(reqCtx, L)
		}))
		return 1
	case "state_get":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqStateGet(reqCtx, L)
		}))
		return 1
	case "state_set":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqStateSet(reqCtx, L)
		}))
		return 1
	case "state_set_all":
		L.Push(L.NewFunction(func(L *lua.LState) int {
			return reqStateSetAll(reqCtx, L)
		}))
		return 1
	}

	return 0
}

func GetUserClaim(ctx *gin.Context, signer *signer.Signer) (*signer.SpaceClaim, error) {
	claim, err := signer.ParseSpace(ctx.GetHeader("Authorization"))
	if err != nil {
		return nil, err
	}

	return claim, nil
}

func getSpaceClaim(reqCtx *luaHttpRequestContext) (*signer.SpaceClaim, error) {
	if reqCtx.spaceClaim != nil {
		return reqCtx.spaceClaim, nil
	}

	claim, err := GetUserClaim(reqCtx.ctx, reqCtx.sig)
	if err != nil {
		return nil, err
	}
	if claim.SpaceId != reqCtx.spaceId {
		return nil, errors.New("invalid space id")
	}
	reqCtx.spaceClaim = claim
	return claim, nil
}

func reqAbort(reqCtx *luaHttpRequestContext, _ *lua.LState) int {
	reqCtx.ctx.Abort()
	return 0
}

func reqAbortWithStatus(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	code := L.CheckInt(1)
	reqCtx.ctx.AbortWithStatus(code)
	return 0
}

func reqAbortWithStatusJSON(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	code := L.CheckInt(1)
	jsonTbl := L.CheckTable(2)
	jsonObj := luaplus.TableToMap(L, jsonTbl)
	reqCtx.ctx.AbortWithStatusJSON(code, jsonObj)
	return 0
}

func reqClientIP(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	L.Push(lua.LString(reqCtx.ctx.ClientIP()))
	return 1
}

func reqContentType(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	L.Push(lua.LString(reqCtx.ctx.ContentType()))
	return 1
}

func reqCookie(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	name := L.CheckString(1)
	value, err := reqCtx.ctx.Cookie(name)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(value))
	return 1
}

func reqGetClaim(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	claim, err := getSpaceClaim(reqCtx)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	claimTable, err := luaplus.StructToTable(L, claim)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(claimTable)
	return 1
}

func reqGetUserId(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	claim, err := getSpaceClaim(reqCtx)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	userId := claim.UserId
	L.Push(lua.LNumber(userId))
	L.Push(lua.LNil)
	return 2
}

func reqData(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	code := L.CheckInt(1)
	contentType := L.CheckString(2)
	data := []byte(L.CheckString(3))
	reqCtx.ctx.Data(code, contentType, data)
	return 0
}

func reqDefaultQuery(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	key := L.CheckString(1)
	defaultValue := L.CheckString(2)
	L.Push(lua.LString(reqCtx.ctx.DefaultQuery(key, defaultValue)))
	return 1
}

func reqDefaultPostForm(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	key := L.CheckString(1)
	defaultValue := L.CheckString(2)
	L.Push(lua.LString(reqCtx.ctx.DefaultPostForm(key, defaultValue)))
	return 1
}

func reqFullPath(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	L.Push(lua.LString(reqCtx.ctx.FullPath()))
	return 1
}

func reqGetHeader(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	key := L.CheckString(1)
	L.Push(lua.LString(reqCtx.ctx.GetHeader(key)))
	return 1
}

func reqGetQuery(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	key := L.CheckString(1)
	value, exists := reqCtx.ctx.GetQuery(key)
	if !exists {
		L.Push(lua.LNil)
		L.Push(lua.LBool(false))
		return 2
	}
	L.Push(lua.LString(value))
	L.Push(lua.LBool(true))
	return 2
}

func reqGetPostForm(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	key := L.CheckString(1)
	value, exists := reqCtx.ctx.GetPostForm(key)
	if !exists {
		L.Push(lua.LNil)
		L.Push(lua.LBool(false))
		return 2
	}
	L.Push(lua.LString(value))
	L.Push(lua.LBool(true))
	return 2
}

func reqParam(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	key := L.CheckString(1)
	L.Push(lua.LString(reqCtx.ctx.Param(key)))
	return 1
}

func reqRedirect(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	code := L.CheckInt(1)
	location := L.CheckString(2)
	reqCtx.ctx.Redirect(code, location)
	return 0
}

func reqRemoteIP(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	L.Push(lua.LString(reqCtx.ctx.ClientIP()))
	return 1
}

func reqJSON(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	code := L.CheckInt(1)
	jsonTbl := L.CheckTable(2)
	jsonObj := luaplus.TableToMap(L, jsonTbl)
	reqCtx.ctx.JSON(code, jsonObj)
	return 0
}

func reqJSONArray(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	code := L.CheckInt(1)
	jsonTbl := L.CheckTable(2)
	jsonObj := luaplus.TableToArray(L, jsonTbl)
	reqCtx.ctx.JSON(code, jsonObj)
	return 0
}

func reqHTML(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	code := L.CheckInt(1)
	name := L.CheckString(2)
	dataTbl := L.CheckTable(3)
	dataObj := luaplus.TableToMap(L, dataTbl)
	reqCtx.ctx.HTML(code, name, dataObj)
	return 0
}

func reqString(reqCtx *luaHttpRequestContext, L *lua.LState) int {
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
	reqCtx.ctx.String(code, format, values...)
	return 0
}

func reqSetCookie(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	name := L.CheckString(1)
	value := L.CheckString(2)
	maxAge := L.CheckInt(3)
	path := L.CheckString(4)
	domain := L.CheckString(5)
	secure := L.CheckBool(6)
	httpOnly := L.CheckBool(7)
	reqCtx.ctx.SetCookie(name, value, maxAge, path, domain, secure, httpOnly)
	return 0
}

func reqStatus(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	code := L.CheckInt(1)
	reqCtx.ctx.Status(code)
	return 0
}

func reqHeader(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	key := L.CheckString(1)
	value := L.CheckString(2)
	reqCtx.ctx.Header(key, value)
	return 0
}

func reqBindJSON(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	var obj map[string]any
	err := reqCtx.ctx.BindJSON(&obj)
	if err != nil {
		return pushError(L, err)
	}
	result := luaplus.MapToTable(L, obj)
	L.Push(result)
	return 1
}

func reqBindHeader(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	var obj map[string]any
	err := reqCtx.ctx.BindHeader(&obj)
	if err != nil {
		return pushError(L, err)
	}
	result := luaplus.MapToTable(L, obj)
	L.Push(result)
	return 1
}

func reqBindQuery(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	var obj map[string]any
	err := reqCtx.ctx.BindQuery(&obj)
	if err != nil {
		return pushError(L, err)
	}
	result := luaplus.MapToTable(L, obj)
	L.Push(result)
	return 1
}

func reqGetRawData(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	data, err := reqCtx.ctx.GetRawData()
	if err != nil {
		return pushError(L, err)
	}
	L.Push(lua.LString(string(data)))
	return 1
}

func reqFormFile(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	name := L.CheckString(1)
	file, err := reqCtx.ctx.FormFile(name)
	if err != nil {
		return pushError(L, err)
	}
	fileTable := L.NewTable()
	L.SetField(fileTable, "filename", lua.LString(file.Filename))
	L.SetField(fileTable, "size", lua.LNumber(file.Size))
	L.Push(fileTable)
	return 1
}

func reqGetQueryMap(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	key := L.CheckString(1)
	values, exists := reqCtx.ctx.GetQueryMap(key)
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

func reqGetQueryArray(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	key := L.CheckString(1)
	values, exists := reqCtx.ctx.GetQueryArray(key)
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

func reqGetPostFormMap(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	key := L.CheckString(1)
	values, exists := reqCtx.ctx.GetPostFormMap(key)
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

func reqGetPostFormArray(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	key := L.CheckString(1)
	values, exists := reqCtx.ctx.GetPostFormArray(key)
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

func reqStateKeys(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	table := L.NewTable()
	for key := range reqCtx.ctx.Keys {
		L.SetField(table, key, lua.LString(key))
	}
	L.Push(table)
	return 1
}

func reqStateGet(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	key := L.CheckString(1)
	value := reqCtx.ctx.Keys[key]
	if value == nil {
		L.Push(lua.LNil)
		return 1
	}
	lvalue := luaplus.GoTypeToLuaType(L, value)
	L.Push(lvalue)
	return 1
}

func reqStateSet(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	key := L.CheckString(1)
	value := L.CheckAny(2)
	reqCtx.ctx.Set(key, value)
	return 0
}

func reqStateSetAll(reqCtx *luaHttpRequestContext, L *lua.LState) int {
	data := L.CheckTable(1)
	data.ForEach(func(key, value lua.LValue) {
		gvalue := luaplus.LuaTypeToGoType(L, value)
		reqCtx.ctx.Set(key.String(), gvalue)
	})
	return 0
}

func reqSSEvent(reqCtx *luaHttpRequestContext, L *lua.LState) int {
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
		msgValue = luaplus.TableToMap(L, message.(*lua.LTable))
	default:
		msgValue = message.String()
	}

	reqCtx.ctx.SSEvent(name, msgValue)
	return 0
}

func HttpModule(app xtypes.App, spaceId int64, L *lua.LState, ctx *gin.Context) *lua.LUserData {
	return NewHttpRequestContext(L, app, spaceId, ctx)
}
