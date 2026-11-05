package emtyctx

import (
	"errors"

	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
)

var Instance = &EmptyCtx{}

type EmptyCtx struct{}

func (e *EmptyCtx) ListActions() ([]string, error) {
	return []string{}, nil
}

func (e *EmptyCtx) ExecuteAction(name string, params lazydata.LazyData) (any, error) {
	return nil, errors.New("unknown action")
}
