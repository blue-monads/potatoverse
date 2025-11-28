package engine

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/blue-monads/turnix/backend/utils/libx"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type Runtime struct {
	activeExecs     map[int64]xtypes.Executor
	activeExecsLock sync.RWMutex

	builders map[string]xtypes.ExecutorBuilder

	parent *Engine
}

func (r *Runtime) cleanupExecs() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for range ticker.C {
		qq.Println("@cleanup_execs/1")

		r.activeExecsLock.RLock()
		for _, e := range r.activeExecs {
			e.Cleanup()
		}

		r.activeExecsLock.RUnlock()
	}
}

func (r *Runtime) GetDebugData() map[int64]any {

	resp := make(map[int64]any)

	r.activeExecsLock.RLock()
	defer r.activeExecsLock.RUnlock()

	for id, e := range r.activeExecs {
		resp[id] = e.GetDebugData()
	}

	return resp

}

func (r *Runtime) ClearExecs(spaceIds ...int64) {
	r.activeExecsLock.Lock()
	defer r.activeExecsLock.Unlock()
	for _, spaceId := range spaceIds {
		delete(r.activeExecs, spaceId)
	}
}

func (r *Runtime) GetExec(spaceKey string, installedId, pVersionId, spaceid int64) (xtypes.Executor, error) {
	r.activeExecsLock.RLock()
	e := r.activeExecs[spaceid]
	r.activeExecsLock.RUnlock()
	if e != nil {
		return e, nil
	}

	wd := path.Join(r.parent.workingFolder, "work_dir", spaceKey, fmt.Sprintf("%d", pVersionId))

	os.MkdirAll(wd, 0755)

	space, err := r.parent.db.GetSpaceOps().GetSpace(spaceid)
	if err != nil {
		return nil, errors.New("space not found")
	}

	builder, ok := r.builders[space.ExecutorType]
	if !ok {
		return nil, errors.New("executor builder not found")
	}

	rfs, err := os.OpenRoot(wd)
	if err != nil {
		return nil, err
	}

	e, err = builder.Build(&xtypes.ExecutorBuilderOption{
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

	r.activeExecsLock.Lock()
	r.activeExecs[spaceid] = e
	r.activeExecsLock.Unlock()

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

		err := e.HandleHttp(xtypes.HttpExecution{
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
