package luaz

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/blue-monads/turnix/backend/engine/executors/luaz/binds"
	"github.com/blue-monads/turnix/backend/utils/kosher"
	"github.com/blue-monads/turnix/backend/utils/luaplus"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/lazydata"
	"github.com/cjoudrey/gluahttp"
	"github.com/gin-gonic/gin"

	lua "github.com/yuin/gopher-lua"
	luaJson "layeh.com/gopher-json"
)

/*

- implement cleanup loop
- use closer properly

*/

var luaHttpClient = &http.Client{}

type CloseItem struct {
	Closer func() error
	Id     uint16
}

type LuaH struct {
	counter uint16
	parent  *LuazExecutor
	closers []CloseItem
	L       *lua.LState
}

func (l *LuaH) AddCloser(closer func() error) uint16 {
	l.counter++
	l.closers = append(l.closers, CloseItem{Closer: closer, Id: l.counter})
	return l.counter
}

func (l *LuaH) RemoveCloser(id uint16) {
	for i := range l.closers {
		l.closers[i] = CloseItem{Closer: nil, Id: 0}
	}
}

func (l *LuaH) Close() error {
	for _, c := range l.closers {
		if c.Closer != nil {
			err := c.Closer()
			qq.Println("@close/1", "closer", c.Id, "error", err)
		}
	}

	l.closers = l.closers[:0]

	return nil
}

func (l *LuaH) logger() *slog.Logger {
	return l.parent.handle.Logger
}

type LuaContextOptions struct {
	HttpContext *gin.Context
	Params      map[string]string
	HandlerName string
}

func (l *LuaH) HandleHTTP(ctx *gin.Context, handlerName string, params map[string]string) error {
	ctxt := l.L.NewTable()

	lh := l

	l.logger().Info("handling http", "handler", handlerName, "params", params)

	var reqCtx *lua.LUserData

	l.L.SetFuncs(ctxt, map[string]lua.LGFunction{
		"request": func(L *lua.LState) int {
			app := l.parent.parent.app
			spaceId := l.parent.handle.SpaceId

			if reqCtx == nil {
				reqCtx = binds.HttpModule(app, spaceId, L, ctx)
				L.Push(reqCtx)
				return 1
			}

			L.Push(reqCtx)
			return 1
		},
		"type": func(l *lua.LState) int {
			l.Push(lua.LString("http"))
			return 1
		},
		"param": func(l *lua.LState) int {
			key := l.CheckString(1)
			l.Push(lua.LString(params[key]))
			return 1
		},

		"get_user_claim": func(l *lua.LState) int {
			claim, err := binds.GetUserClaim(ctx, lh.parent.parent.app.Signer())
			if err != nil {
				return luaplus.PushError(l, err)
			}
			claimTable, err := luaplus.StructToTable(l, claim)
			if err != nil {
				return luaplus.PushError(l, err)
			}
			l.Push(claimTable)
			return 1
		},

		"get_header": func(l *lua.LState) int {
			key := l.CheckString(1)
			l.Push(lua.LString(ctx.GetHeader(key)))
			return 1
		},

		"get_json": func(l *lua.LState) int {

			var target any

			err := ctx.BindJSON(&target)
			if err != nil {
				return luaplus.PushError(l, err)
			}

			result := luaplus.GoTypeToLuaType(l, target)
			l.Push(result)
			l.Push(lua.LNil)
			return 2

		},

		"set_json": func(l *lua.LState) int {
			code := l.CheckInt(1)
			target := l.CheckTable(2)
			targetMap := luaplus.TableToMap(l, target)
			ctx.JSON(code, targetMap)
			return 0
		},
	})

	l.logger().Info("ctxt")

	err := callHandler(l, ctxt, handlerName)
	if err != nil {
		return err
	}

	return nil

}

var EmptyLazyData = lazydata.LazyDataBytes([]byte("{}"))

func (l *LuaH) HandleAction(event *xtypes.ActionEvent) error {

	ctxt := l.L.NewTable()

	l.L.SetFuncs(ctxt, map[string]lua.LGFunction{

		"type": func(L *lua.LState) int {
			L.Push(lua.LString("action"))
			return 1
		},
		"param": func(L *lua.LState) int {
			key := L.CheckString(1)
			L.Push(lua.LString(event.Params[key]))
			return 1
		},

		"get_inner_payload": func(L *lua.LState) int {
			result, err := event.Request.ExecuteAction("get_payload", EmptyLazyData)
			if err != nil {
				return luaplus.PushError(L, err)
			}
			resultTable := luaplus.GoTypeToLuaType(L, result)
			L.Push(resultTable)
			L.Push(lua.LNil)
			return 2
		},

		"get_inner_value": func(L *lua.LState) int {
			field := L.CheckString(1)
			paramPayload := fmt.Sprintf(`{"path": "%s"}`, field)
			paramLazyData := lazydata.LazyDataBytes(kosher.Byte(paramPayload))
			result, err := event.Request.ExecuteAction("get_value", paramLazyData)
			if err != nil {
				return luaplus.PushError(L, err)
			}
			resultTable := luaplus.GoTypeToLuaType(L, result)
			L.Push(resultTable)
			L.Push(lua.LNil)
			return 2
		},

		"execute": func(L *lua.LState) int {
			actionName := L.CheckString(1)
			params := L.CheckTable(2)

			paramsLazyData := binds.NewLuaLazyData(L, params)

			result, err := event.Request.ExecuteAction(actionName, paramsLazyData)
			if err != nil {
				return luaplus.PushError(L, err)
			}
			resultTable := luaplus.GoTypeToLuaType(L, result)
			L.Push(resultTable)
			L.Push(lua.LNil)
			return 2
		},

		"list_methods": func(L *lua.LState) int {
			actions, err := event.Request.ListActions()
			if err != nil {
				L.Push(lua.LNil)
				return 1
			}
			table := L.NewTable()
			for _, action := range actions {
				L.Push(lua.LString(action))
			}

			L.Push(table)

			return 1
		},
	})

	method := fmt.Sprintf("on_%s", event.EventType)

	err := callHandler(l, ctxt, method)
	if err != nil {
		return err
	}

	return nil

}

func callHandler(l *LuaH, ctable *lua.LTable, handlerName string) error {
	handler := l.L.GetGlobal(handlerName)
	if handler == lua.LNil {
		qq.Println("@callHandler/1", "handler not found", handlerName)
		return errors.New("handler not found")
	}

	if handler == nil {
		qq.Println("@callHandler/2", "handler is nil", handlerName)
		return errors.New("handler is nil")
	}

	qq.Println("@callHandler/3", "handler", handler.String())

	l.L.Push(handler)

	qq.Println("@callHandler/4", "handler pushed")

	l.L.Push(ctable)

	qq.Println("@callHandler/5", "ctable pushed")

	l.L.Call(1, 0)

	qq.Println("@callHandler/6", "handler called")

	return nil

}

func (l *LuaH) registerModules() error {
	installId := l.parent.handle.InstalledId
	spaceId := l.parent.handle.SpaceId
	app := l.parent.parent.app
	packageVersionId := l.parent.handle.PackageVersionId

	l.L.PreloadModule("pmcp", binds.BindMCP)
	l.L.PreloadModule("potato", binds.PotatoModule(app, installId, packageVersionId, spaceId))
	l.L.PreloadModule("phttp", gluahttp.NewHttpModule(luaHttpClient).Loader)
	l.L.PreloadModule("json", luaJson.Loader)

	return nil
}
