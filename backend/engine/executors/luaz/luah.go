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

func (l *LuaH) Handle(ctx *gin.Context, handlerName string, params map[string]string) error {
	handler := l.L.GetGlobal(handlerName)
	ctxt := l.L.NewTable()
	pp.Println("@LuaH/handle/1")

	if handler == lua.LNil {
		pp.Println("@LuaH/handle/2", handlerName, params)
		pp.Println("@LuaH/handle/3", handler)

		l.logger().Error("handler not found", "handler", handlerName)
		// Debug: check if some known functions exist
		testHandler := l.L.GetGlobal("get_category_page")
		if testHandler != lua.LNil {
			l.logger().Error("get_category_page exists but handler not found", "handler", handlerName)
		} else {
			l.logger().Error("no functions found in lua state")
		}

		return errors.New("handler not found")
	}

	if handler == nil {
		pp.Println("@LuaH/handle/4")
	}

	pp.Println("@LuaH/handle/5")

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

	pp.Println("@LuaH/handle/6")

	l.L.Push(handler)
	pp.Println("@LuaH/handle/7")
	l.L.Push(ctxt)
	pp.Println("@LuaH/handle/8")
	l.L.Call(1, 0)
	pp.Println("@LuaH/handle/9")

	return nil

}

func (l *LuaH) registerModules() error {
	l.L.PreloadModule("pkv", binds.BindsKV(l.parent.handle.SpaceId, l.parent.handle))
	l.L.PreloadModule("kv", binds.BindsKV(l.parent.handle.RootSpaceId, l.parent.handle))
	l.L.PreloadModule("mcp", binds.BindMCP())
	l.L.PreloadModule("ufs", binds.UfsModule(l.parent.handle))
	l.L.PreloadModule("core", binds.CoreModule(l.parent.handle))

	return nil
}
