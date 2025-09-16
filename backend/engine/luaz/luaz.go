package luaz

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
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
		MinSize:     10,
		MaxSize:     20,
		MaxOnFlight: 50,
		Ttl:         time.Hour,
		InitFn: func() (*LuaH, error) {

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
	pp.Println("@cleanup/2")
	l.pool.CleanupExpiredStates()
	pp.Println("@cleanup/3")
}

type HttpEvent struct {
	HandlerName string
	Params      map[string]string
	Request     *gin.Context
}

func (l *Luaz) Handle(event HttpEvent) {
	pp.Println("@handle/1")

	lh, err := l.pool.Get()
	if err != nil {
		pp.Println("@handle/1.1", err)
		httpx.WriteErr(event.Request, err)
		return
	}

	if lh == nil {
		pp.Println("@handle/1.2", "lh is nil")
		httpx.WriteErr(event.Request, errors.New("Could not get lua state"))
		return
	}

	pp.Println("@handle/2", event.HandlerName, event.Params)

	lh.Handle(event.Request, event.HandlerName, event.Params)

	pp.Println("@handle/3")

	l.pool.Put(lh)

}

func (l *Luaz) GetDebugData() map[string]any {
	return l.pool.GetDebugData()
}
