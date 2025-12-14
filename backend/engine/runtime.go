package engine

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sync"

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

func (r *Runtime) ExecHttp(nsKey string, installedId, packageVersionId, spaceId int64, ctx *gin.Context) error {

	e, err := r.GetExec(nsKey, installedId, packageVersionId, spaceId)
	if err != nil {
		qq.Println("@exec_http/1", "error getting exec", err)
		httpx.WriteErr(ctx, err)
		return err
	}

	if e == nil {
		qq.Println("@exec_http/1", "exec is nil")
		httpx.WriteErr(ctx, errors.New("exec is nil"))
		return errors.New("exec is nil")
	}

	// print stack trace

	err = libx.PanicWrapper(func() {
		subpath := ctx.Param("subpath")

		params := make(map[string]string)

		params["space_id"] = fmt.Sprintf("%d", spaceId)
		params["install_id"] = fmt.Sprintf("%d", installedId)
		params["package_version_id"] = fmt.Sprintf("%d", packageVersionId)
		params["subpath"] = subpath
		params["method"] = ctx.Request.Method

		err := e.HandleHttp(xtypes.HttpExecution{
			HandlerName: "on_http",
			Params:      params,
			Request:     ctx,
		})
		if err != nil {
			qq.Println("@exec_http/2", "error handling http", err)
			panic(err)
		}

	})

	return err

}

func (r *Runtime) ExecEvent(nsKey string, installedId, packageVersionId, spaceId int64, req xtypes.GenericRequest) error {

	e, err := r.GetExec(nsKey, installedId, packageVersionId, spaceId)
	if err != nil {
		qq.Println("@exec_event/1", "error getting exec", err)
		return err
	}

	if e == nil {
		qq.Println("@exec_event/1", "exec is nil")
		return errors.New("exec is nil")
	}

	err = e.HandleEvent(xtypes.EventExecution{
		Type:       "ws",
		ActionName: "",
		Params:     make(map[string]string),
		Request:    nil,
	})

	return err

}
