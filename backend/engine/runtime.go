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
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
	"github.com/k0kubun/pp"
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
		pp.Println("@cleanup_execs/1")

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

func (r *Runtime) GetExec(packageName string, packageId, spaceid int64) (*luaz.Luaz, error) {
	r.execsLock.RLock()
	e := r.execs[packageId]
	r.execsLock.RUnlock()
	if e != nil {
		return e, nil
	}

	source := Code

	if !ByPassPackageCode {
		file, err := r.parent.db.GetPackageFileMetaByPath(packageId, "", "server.lua")
		if err != nil {
			r.parent.app.Logger().Error("error getting package file meta by path", "error", err)
			return nil, err
		}

		sourceBytes, err := r.parent.db.GetPackageFile(packageId, file.ID)
		if err != nil {
			r.parent.app.Logger().Error("error getting package file", "error", err)
			return nil, err
		}

		source = string(sourceBytes)

	}

	e = luaz.New(luaz.Options{
		BuilderOpts: xtypes.BuilderOption{
			App:    r.parent.app,
			Logger: r.parent.app.Logger().With("package_id", packageId),
		},
		Code:          source, // code,
		WorkingFolder: path.Join(r.parent.workingFolder, packageName, fmt.Sprintf("%d", packageId)),
		SpaceId:       spaceid,
		PackageId:     packageId,
	})

	r.execsLock.Lock()
	r.execs[packageId] = e
	r.execsLock.Unlock()

	return e, nil

}

func (r *Runtime) ExecHttpWithErr(packageName string, packageId, spaceId int64, ctx *gin.Context) error {
	return r.ExecHttpWithHandlerAndErr(packageName, packageId, spaceId, "on_http", ctx)
}

func (r *Runtime) ExecHttpWithHandlerAndErr(packageName string, packageId, spaceId int64, handlerName string, ctx *gin.Context) error {

	e, err := r.GetExec(packageName, packageId, spaceId)
	if err != nil {
		pp.Println("@exec_http/1", "error getting exec", err)
		httpx.WriteErr(ctx, err)
		return err
	}

	if e == nil {
		pp.Println("@exec_http/1", "exec is nil")
		httpx.WriteErr(ctx, errors.New("exec is nil"))
		return errors.New("exec is nil")
	}

	// print stack trace

	err = libx.PanicWrapper(func() {
		subpath := ctx.Param("subpath")

		err := e.Handle(luaz.HttpEvent{
			HandlerName: handlerName,
			Params: map[string]string{
				"space_id":   fmt.Sprintf("%d", spaceId),
				"package_id": fmt.Sprintf("%d", packageId),
				"subpath":    subpath,
			},
			Request: ctx,
		})
		if err != nil {
			pp.Println("@exec_http/2", "error handling http", err)
			panic(err)
		}

	})

	if err != nil {
		return err
	}

	return nil

}

func (r *Runtime) ExecHttp(packageName string, packageId, spaceId int64, ctx *gin.Context) {
	err := r.ExecHttpWithErr(packageName, packageId, spaceId, ctx)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

}
