package engine

import (
	"errors"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/blue-monads/turnix/backend/engine/executors/luaz"
	"github.com/blue-monads/turnix/backend/utils/libx"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
)

const Code = `

local db = require("db")
local math = require("math")

function im_cool(a)
	print("I'm cool")
	return a + 1
end


function on_http(ctx)
  print("Hello from lua!", ctx.type())
  local req = ctx.request()

  local rand = math.random(1, 100)

  db.add({
	group = "test",
	key = "test" .. rand,
	value = "test",
  })


  req.json(200, {
	im_cool = im_cool(18),
	message = "Hello from lua! from lua!",
	space_id = ctx.param("space_id"),
	package_id = ctx.param("package_id"),
	subpath = ctx.param("subpath"),
  })

end

`

const ByPassPackageCode = false

type Runtime struct {
	execs     map[int64]*luaz.Luaz
	execsLock sync.RWMutex
	parent    *Engine
}

func (r *Runtime) cleanupExecs() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for range ticker.C {
		qq.Println("@cleanup_execs/1")

		r.execsLock.RLock()
		for _, e := range r.execs {
			e.Cleanup()
		}

		r.execsLock.RUnlock()
	}
}

func (r *Runtime) GetDebugData() map[int64]any {

	resp := make(map[int64]any)

	r.execsLock.RLock()
	defer r.execsLock.RUnlock()

	for id, e := range r.execs {
		resp[id] = e.GetDebugData()
	}

	return resp

}

func (r *Runtime) GetExec(spaceKey string, pVersionId, spaceid int64) (*luaz.Luaz, error) {
	r.execsLock.RLock()
	e := r.execs[spaceid]
	r.execsLock.RUnlock()
	if e != nil {
		return e, nil
	}

	source := Code

	if !ByPassPackageCode {

		sOps := r.parent.db.GetSpaceOps()
		s, err := sOps.GetSpace(spaceid)
		if err != nil {
			return nil, err
		}

		if s.ServerFile == "" {
			s.ServerFile = "server.lua"
		}

		pfops := r.parent.db.GetPackageFileOps()
		packageFile, err := pfops.GetFileContentByPath(pVersionId, "", s.ServerFile)
		if err != nil {
			return nil, err
		}

		source = string(packageFile)
	}

	e = luaz.New(luaz.Options{
		BuilderOpts: xtypes.BuilderOption{
			App:    r.parent.app,
			Logger: r.parent.app.Logger().With("package_id", pVersionId),
		},
		Code:             source, // code,
		WorkingFolder:    path.Join(r.parent.workingFolder, spaceKey, fmt.Sprintf("%d", pVersionId)),
		SpaceId:          spaceid,
		PackageVersionId: 0,
		InstalledId:      0,
	})

	r.execsLock.Lock()
	r.execs[spaceid] = e
	r.execsLock.Unlock()

	return e, nil

}

type ExecuteOptions struct {
	PackageName      string
	PackageVersionId int64
	InstalledId      int64
	SpaceId          int64
	HandlerName      string
	HttpContext      *gin.Context
	Params           map[string]string
}

func (r *Runtime) ExecuteHttp(opts ExecuteOptions) error {

	e, err := r.GetExec(opts.PackageName, opts.InstalledId, opts.SpaceId)
	if err != nil {
		qq.Println("@exec_http/1", "error getting exec", err)
		httpx.WriteErr(opts.HttpContext, err)
		return err
	}

	if e == nil {
		qq.Println("@exec_http/1", "exec is nil")
		httpx.WriteErr(opts.HttpContext, errors.New("exec is nil"))
		return errors.New("exec is nil")
	}

	// print stack trace

	err = libx.PanicWrapper(func() {
		subpath := opts.HttpContext.Param("subpath")

		params := opts.Params
		if params == nil {
			params = make(map[string]string)
		}

		params["space_id"] = fmt.Sprintf("%d", opts.SpaceId)
		params["install_id"] = fmt.Sprintf("%d", opts.InstalledId)
		params["subpath"] = subpath
		params["method"] = opts.HttpContext.Request.Method

		err := e.Handle(luaz.HttpEvent{
			HandlerName: opts.HandlerName,
			Params:      params,
			Request:     opts.HttpContext,
		})
		if err != nil {
			qq.Println("@exec_http/2", "error handling http", err)
			panic(err)
		}

	})

	if err != nil {
		return err
	}

	return nil

}

func (r *Runtime) ExecHttp(packageName string, installedId, spaceId int64, ctx *gin.Context) {
	err := r.ExecuteHttp(ExecuteOptions{
		PackageName: packageName,
		InstalledId: installedId,
		SpaceId:     spaceId,
		HandlerName: "on_http",
		HttpContext: ctx,
	})
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

}
