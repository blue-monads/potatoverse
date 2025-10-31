package luaz

import (
	"errors"
	"log/slog"

	"github.com/blue-monads/turnix/backend/engine/executors/luaz/binds"
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
	lua "github.com/yuin/gopher-lua"
)

type LuaH struct {
	parent  *Luaz
	closers []func() error
	L       *lua.LState
}

func (l *LuaH) Close() error {
	for _, c := range l.closers {
		c()
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

func (l *LuaH) Handle(ctx *gin.Context, handlerName string, params map[string]string) error {
	ctxt := l.L.NewTable()

	l.logger().Info("handling http", "handler", handlerName, "params", params)

	l.L.SetFuncs(ctxt, map[string]lua.LGFunction{
		"request": func(L *lua.LState) int {
			table := binds.HttpModule(l.parent.handle, L, ctx)
			L.Push(table)
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

func callHandler(l *LuaH, ctable *lua.LTable, handlerName string) error {
	handler := l.L.GetGlobal(handlerName)
	if handler == lua.LNil {
		pp.Println("@callHandler/1", "handler not found", handlerName)
		return errors.New("handler not found")
	}

	if handler == nil {
		pp.Println("@callHandler/2", "handler is nil", handlerName)
		return errors.New("handler is nil")
	}

	pp.Println("@callHandler/3", "handler", handler.String())

	l.L.Push(handler)

	pp.Println("@callHandler/4", "handler pushed")

	l.L.Push(ctable)

	pp.Println("@callHandler/5", "ctable pushed")

	l.L.Call(1, 0)

	pp.Println("@callHandler/6", "handler called")

	return nil

}

func (l *LuaH) registerModules() error {
	l.L.PreloadModule("pkv", binds.BindsKV(l.parent.handle.SpaceId, l.parent.handle))
	l.L.PreloadModule("kv", binds.BindsKV(l.parent.handle.RootSpaceId, l.parent.handle))
	l.L.PreloadModule("mcp", binds.BindMCP)
	l.L.PreloadModule("addon", binds.AddOnModule(l.parent.handle))

	return nil
}
