package luaz

import (
	"errors"
	"time"

	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
)

type Luaz struct {
	pool   *LuaStatePool
	handle *xtypes.BuilderOption
}

func New(opts *xtypes.BuilderOption) (*Luaz, error) {

	source := Code

	if !ByPassPackageCode {
		sOps := opts.App.Database().GetSpaceOps()
		s, err := sOps.GetSpace(opts.SpaceId)
		if err != nil {
			return nil, errors.New("space not found")
		}

		if s.ServerFile == "" {
			s.ServerFile = "server.lua"
		}

		pfops := opts.App.Database().GetPackageFileOps()
		packageFile, err := pfops.GetFileContentByPath(opts.PackageVersionId, "", s.ServerFile)

		if err != nil {
			qq.Println("@lua_exec_error", err)
			qq.Println("@package file not found", opts.PackageVersionId, opts.SpaceId, s.ServerFile)
			qq.Println("@space", s)
			return nil, errors.New("package file not found")
		}

		source = string(packageFile)

	}

	lz := &Luaz{
		pool:   nil,
		handle: opts,
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

			err := lh.registerModules()
			if err != nil {
				return nil, err
			}

			err = L.DoString(source)
			if err != nil {
				qq.Println("@lua_exec_error", err)
				return nil, err
			}
			qq.Println("@lua_exec_success", "code length", len(source))

			return lh, nil
		},
	})

	lz.pool = pool

	return lz, nil
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
