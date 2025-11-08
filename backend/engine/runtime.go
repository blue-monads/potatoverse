package engine

import (
	"errors"
	"fmt"
	"os"
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

func (r *Runtime) ClearExecs() {
	r.execsLock.Lock()
	defer r.execsLock.Unlock()
	r.execs = make(map[int64]*luaz.Luaz)
}

func (r *Runtime) GetExec(spaceKey string, installedId, pVersionId, spaceid int64) (*luaz.Luaz, error) {
	r.execsLock.RLock()
	e := r.execs[spaceid]
	r.execsLock.RUnlock()
	if e != nil {
		return e, nil
	}

	wd := path.Join(r.parent.workingFolder, "work_dir", spaceKey, fmt.Sprintf("%d", pVersionId))

	os.MkdirAll(wd, 0755)

	rfs, err := os.OpenRoot(wd)
	if err != nil {
		return nil, err
	}

	e, err = luaz.New(&xtypes.ExecutorBuilderOption{
		App:              r.parent.app,
		Logger:           r.parent.app.Logger().With("package_id", pVersionId),
		WorkingFolder:    wd,
		SpaceId:          spaceid,
		InstalledId:      installedId,
		PackageVersionId: pVersionId,
		FsRoot:           rfs,
	})
	if err != nil {
		return nil, err
	}

	r.execsLock.Lock()
	r.execs[spaceid] = e
	r.execsLock.Unlock()

	return e, nil

}

type ExecuteOptions struct {
	NSKey            string
	PackageVersionId int64
	InstalledId      int64
	SpaceId          int64
	HandlerName      string
	HttpContext      *gin.Context
	Params           map[string]string
}

func (r *Runtime) ExecuteHttp(opts ExecuteOptions) error {

	e, err := r.GetExec(opts.NSKey, opts.InstalledId, opts.PackageVersionId, opts.SpaceId)
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
		params["package_version_id"] = fmt.Sprintf("%d", opts.PackageVersionId)
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

func (r *Runtime) ExecHttp(nsKey string, installedId, packageVersionId, spaceId int64, ctx *gin.Context) {
	err := r.ExecuteHttp(ExecuteOptions{
		NSKey:            nsKey,
		InstalledId:      installedId,
		PackageVersionId: packageVersionId,
		SpaceId:          spaceId,
		HandlerName:      "on_http",
		HttpContext:      ctx,
		Params:           make(map[string]string),
	})
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

}
