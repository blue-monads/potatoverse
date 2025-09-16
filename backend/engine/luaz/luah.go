package luaz

import (
	"github.com/blue-monads/turnix/backend/engine/luaz/binds"
	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
)

const code = `
function handler(ctx)
  print("Hello from lua!", ctx.type())
end

`

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

func (l *LuaH) Handle(ctx *gin.Context, handlerName string, params map[string]string) {
	handler := l.L.GetGlobal(handlerName)
	ctxt := l.L.NewTable()

	if handler == lua.LNil {
		l.parent.logger.Error("handler not found", "handler", handlerName)
		return
	}

	l.L.SetFuncs(ctxt, map[string]lua.LGFunction{
		"request": func(l *lua.LState) int {
			table := binds.HttpModule(l, ctx)
			l.Push(table)
			return 1
		},
		"type": func(l *lua.LState) int {
			l.Push(lua.LString("http"))
			return 1
		},
	})

	l.L.Push(handler)
	l.L.Push(ctxt)
	l.L.Call(1, 0)

}

func (l *LuaH) registerModules() error {
	l.L.PreloadModule("fs", binds.FsModule(l.parent.root))

	return nil
}
