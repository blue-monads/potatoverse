package luaz

import (
	"log/slog"
	"os"
	"time"

	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
)

type Luaz struct {
	pool   *LuaStatePool
	app    xtypes.App
	logger *slog.Logger
	root   *os.Root
}

type Options struct {
	BuilderOpts   xtypes.BuilderOption
	Code          string
	WorkingFolder string
}

func New(opts Options) *Luaz {
	lz := &Luaz{
		pool: nil,
	}

	pool := NewLuaStatePool(LuaStatePoolOptions{
		MinSize: 10,
		MaxSize: 20,
		Ttl:     time.Hour,
		InitFn: func() (*LuaH, error) {

			if opts.Code == "" {
				opts.Code = code
			}

			L := lua.NewState()
			err := L.DoString(opts.Code)
			if err != nil {
				return nil, err
			}

			lh := &LuaH{
				parent:  lz,
				L:       L,
				closers: []func() error{},
			}

			err = lh.registerModules()
			if err != nil {
				return nil, err
			}

			return lh, nil
		},
	})

	lz.pool = pool

	return lz
}

func (l *Luaz) Cleanup() {
	l.pool.CleanupExpiredStates()
}

func (l *Luaz) Handle(ctx *gin.Context, handlerName string) {

	lh, err := l.pool.Get()
	if err != nil {
		return
	}

	lh.Handle(ctx, handlerName, map[string]string{})

}
