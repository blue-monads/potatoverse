package luaz

import (
	"log/slog"

	"github.com/blue-monads/turnix/backend/engine/executors/luaz/binds"
	"github.com/gin-gonic/gin"
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

func (l *LuaH) Handle(ctx *gin.Context, handlerName string, params map[string]string) {
	handler := l.L.GetGlobal(handlerName)
	ctxt := l.L.NewTable()

	if handler == lua.LNil {
		l.logger().Error("handler not found", "handler", handlerName)
		return
	}

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

	l.L.Push(handler)
	l.L.Push(ctxt)
	l.L.Call(1, 0)

}

func (l *LuaH) registerModules() error {
	l.L.PreloadModule("kv", binds.BindsKV(l.parent.handle))
	l.L.PreloadModule("ufs", binds.UfsModule(l.parent.handle))

	return nil
}
