package luaz

import (
	"errors"
	"fmt"
	"time"

	"github.com/blue-monads/potatoverse/backend/engine/executors/luaz/binds"
	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	lua "github.com/yuin/gopher-lua"
)

func init() {
	registry.RegisterExecutorBuilderFactory("luaz", BuildLuazExecutorBuilder)
}

func BuildLuazExecutorBuilder(app xtypes.App) (xtypes.ExecutorBuilder, error) {
	binds := binds.PotatoBindable(app)

	return &LuazExecutorBuilder{app: app, binds: binds}, nil
}

type LuazExecutorBuilder struct {
	app   xtypes.App
	binds map[string]map[string]lua.LGFunction
}

func (b *LuazExecutorBuilder) Name() string {
	return "luaz"
}

func (b *LuazExecutorBuilder) Icon() string {
	return "luaz"
}

func (b *LuazExecutorBuilder) Build(opt *xtypes.ExecutorBuilderOption) (xtypes.Executor, error) {

	source := Code

	if !ByPassPackageCode {
		sOps := b.app.Database().GetSpaceOps()
		s, err := sOps.GetSpace(opt.SpaceId)
		if err != nil {
			return nil, errors.New("space not found")
		}

		if s.ServerFile == "" {
			s.ServerFile = "server.lua"
		}

		pfops := b.app.Database().GetPackageFileOps()
		packageFile, err := pfops.GetFileContentByPath(opt.PackageVersionId, "", s.ServerFile)

		if err != nil {
			qq.Println("@script file load error", err)
			qq.Println("@package file not found", opt.PackageVersionId, opt.SpaceId, s.ServerFile)
			qq.Println("@space", s)
			return nil, errors.New("package file not found")
		}

		source = string(packageFile)

	}

	ex := &LuazExecutor{
		parent: b,
		handle: opt,
	}

	pool := NewLuaStatePool(LuaStatePoolOptions{
		MinSize:     10,
		MaxSize:     20,
		MaxOnFlight: 50,
		Ttl:         time.Hour,
		InitFn: func() (*LuaH, error) {

			L := lua.NewState()

			lh := &LuaH{
				parent:  ex,
				L:       L,
				closers: make([]CloseItem, 0, 4),
				counter: 0,
			}

			err := lh.registerModules()
			if err != nil {
				return nil, err
			}

			err = L.DoString(source)
			if err != nil {
				qq.Println("@lua_exec_error", err)
				qq.Println("@lua_exec_error_error", err.Error())
				fmt.Println("--------------------------------")
				fmt.Println(source)
				fmt.Println("--------------------------------")

				return nil, err
			}
			qq.Println("@lua_exec_success", "code length", len(source))

			return lh, nil
		},
	})

	ex.pool = pool

	return ex, nil

}
