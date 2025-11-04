package luaz

import (
	"errors"
	"os"
	"time"

	"github.com/blue-monads/turnix/backend/engine/executors"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
)

type Luaz struct {
	pool   *LuaStatePool
	handle *executors.EHandle
}

type Options struct {
	BuilderOpts      xtypes.BuilderOption
	Code             string
	WorkingFolder    string
	SpaceId          int64
	PackageVersionId int64
	InstalledId      int64
}

func New(opts Options) *Luaz {

	os.MkdirAll(opts.WorkingFolder, 0755)

	rfs, err := os.OpenRoot(opts.WorkingFolder)
	if err != nil {
		panic(err)
	}

	lz := &Luaz{
		pool: nil,
		handle: &executors.EHandle{
			Logger:           opts.BuilderOpts.Logger,
			FsRoot:           rfs,
			SpaceId:          opts.SpaceId,
			PackageVersionId: opts.PackageVersionId,
			InstalledId:      opts.InstalledId,
			App:              opts.BuilderOpts.App,
			Database:         opts.BuilderOpts.App.Database(),
		},
	}

	pool := NewLuaStatePool(LuaStatePoolOptions{
		MinSize:     10,
		MaxSize:     20,
		MaxOnFlight: 50,
		Ttl:         time.Hour,
		InitFn: func() (*LuaH, error) {

			L := lua.NewState()

			lh := &LuaH{
				parent:  lz,
				L:       L,
				closers: []func() error{},
			}

			err = lh.registerModules()
			if err != nil {
				return nil, err
			}

			err := L.DoString(opts.Code)
			if err != nil {
				qq.Println("@lua_exec_error", err)
				return nil, err
			}
			qq.Println("@lua_exec_success", "code length", len(opts.Code))

			return lh, nil
		},
	})

	lz.pool = pool

	return lz
}

func (l *Luaz) Cleanup() {
	qq.Println("@cleanup/2")
	l.pool.CleanupExpiredStates()
	qq.Println("@cleanup/3")
}

type HttpEvent struct {
	HandlerName string
	Params      map[string]string
	Request     *gin.Context
}

func (l *Luaz) Handle(event HttpEvent) error {
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

	err = lh.Handle(event.Request, event.HandlerName, event.Params)
	if err != nil {
		return err
	}

	qq.Println("@handle/3")

	l.pool.Put(lh)

	return nil

}

func (l *Luaz) GetDebugData() map[string]any {
	return l.pool.GetDebugData()
}

const HandlersReference = `


function on_http(ctx)
	print("@on_http", ctx.type())
end

function on_ws_room(ctx)
	print("@on_ws_room", ctx.type())
end

function on_rmcp(ctx)
	print("@on_rmcp", ctx.type())
end




`
