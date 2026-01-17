package executor

import lua "github.com/yuin/gopher-lua"

type EventHubExecutor struct {
	l *lua.LState
}
