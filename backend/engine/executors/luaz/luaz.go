package luaz

import (
	"errors"

	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
)

var _ xtypes.Executor = (*LuazExecutor)(nil)

type LuazExecutor struct {
	parent *LuazExecutorBuilder
	pool   *LuaStatePool
	handle *xtypes.ExecutorBuilderOption
}

func (l *LuazExecutor) Cleanup() {
	qq.Println("@cleanup/2")
	l.pool.CleanupExpiredStates()
	qq.Println("@cleanup/3")
}

func (l *LuazExecutor) HandleHttp(event *xtypes.HttpExecution) error {
	qq.Println("@handle/1")

	lh, err := l.pool.Get()
	if err != nil {
		qq.Println("@handle/1.1", err)
		httpx.WriteErr(event.Request, err)
		return err
	}

	if lh == nil {
		qq.Println("@handle/1.2", "lh is nil")
		httpx.WriteErr(event.Request, errors.New("Could not get lua state"))
		return errors.New("Could not get lua state")
	}

	qq.Println("@handle/2", event.HandlerName, event.Params)

	err = lh.HandleHTTP(event.Request, event.HandlerName, event.Params)
	if err != nil {
		return err
	}

	qq.Println("@handle/3")

	l.pool.Put(lh)

	return nil

}

func (l *LuazExecutor) HandleAction(event *xtypes.ActionExecution) error {

	lh, err := l.pool.Get()
	if err != nil {
		return err
	}

	if lh == nil {
		return errors.New("Could not get lua state")
	}

	err = lh.HandleAction(event)
	if err != nil {
		return err
	}

	l.pool.Put(lh)

	return nil

}

func (l *LuazExecutor) GetDebugData() map[string]any {
	return l.pool.GetDebugData()
}
