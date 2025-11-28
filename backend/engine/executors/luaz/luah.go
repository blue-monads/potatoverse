package luaz

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/blue-monads/turnix/backend/engine/executors/luaz/binds"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/cjoudrey/gluahttp"
	"github.com/gin-gonic/gin"

	lua "github.com/yuin/gopher-lua"
)

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

	l.logger().Info("handling http", "handler", handlerName, "params", params)

	l.L.SetFuncs(ctxt, map[string]lua.LGFunction{
		"request": func(L *lua.LState) int {
			app := l.parent.parent.app
			spaceId := l.parent.handle.SpaceId

			reqCtx := binds.HttpModule(app, spaceId, L, ctx)
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
	})

	l.logger().Info("ctxt")

	err := callHandler(l, ctxt, handlerName)
	if err != nil {
		return err
	}

	return nil

}

func (l *LuaH) HandleEvent(event xtypes.EventExecution) error {

	ctxt := l.L.NewTable()

	l.L.SetFuncs(ctxt, map[string]lua.LGFunction{
		"request": func(L *lua.LState) int {
			app := l.parent.parent.app
			installId := l.parent.handle.InstalledId
			spaceId := l.parent.handle.SpaceId
			greq := event.Request

			genCtxUserdata := binds.GenericContextModule(app, installId, spaceId, L, greq)
			L.Push(genCtxUserdata)
			return 1
		},
		"type": func(L *lua.LState) int {
			L.Push(lua.LString("event"))
			return 1
		},
		"param": func(L *lua.LState) int {
			key := L.CheckString(1)
			L.Push(lua.LString(event.Params[key]))
			return 1
		},
	})

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

	l.L.PreloadModule("pmcp", binds.BindMCP)
	l.L.PreloadModule("potato", binds.PotatoModule(app, installId, spaceId))
	l.L.PreloadModule("phttp", gluahttp.NewHttpModule(luaHttpClient).Loader)

	return nil
}
