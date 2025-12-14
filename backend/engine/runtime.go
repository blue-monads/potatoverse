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

type RunningExec struct {
	Executor         xtypes.Executor
	SpaceId          int64
	PackageVersionId int64
	InstalledId      int64
}

type Runtime struct {
	activeExecs     map[int64]*RunningExec
	activeExecsLock sync.RWMutex

	builders map[string]xtypes.ExecutorBuilder

	parent *Engine
}

func (r *Runtime) GetDebugData() map[int64]any {

	resp := make(map[int64]any)

	r.activeExecsLock.RLock()
	defer r.activeExecsLock.RUnlock()

	for id, e := range r.activeExecs {
		resp[id] = e.Executor.GetDebugData()
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

func (r *Runtime) GetExec(spaceid int64) (*RunningExec, error) {
	r.activeExecsLock.RLock()
	e := r.activeExecs[spaceid]
	r.activeExecsLock.RUnlock()
	if e != nil {
		return e, nil
	}

	space, err := r.parent.db.GetSpaceOps().GetSpace(spaceid)
	if err != nil {
		return nil, errors.New("space not found")
	}

	pkg, err := r.parent.db.GetPackageInstallOps().GetPackage(space.InstalledId)
	if err != nil {
		return nil, errors.New("package not found")
	}

	spaceKey := space.NamespaceKey

	wd := path.Join(r.parent.workingFolder, "work_dir", spaceKey, fmt.Sprintf("%d", pkg.ActiveInstallID))

	os.MkdirAll(wd, 0755)

	builder, ok := r.builders[space.ExecutorType]
	if !ok {
		return nil, errors.New("executor builder not found")
	}

	rfs, err := os.OpenRoot(wd)
	if err != nil {
		return nil, err
	}

	innerExec, err := builder.Build(&xtypes.ExecutorBuilderOption{
		Logger:           r.parent.app.Logger().With("package_id", pkg.ActiveInstallID),
		WorkingFolder:    wd,
		SpaceId:          spaceid,
		InstalledId:      pkg.ID,
		PackageVersionId: pkg.ActiveInstallID,
		FsRoot:           rfs,
	})

	e = &RunningExec{
		Executor:         innerExec,
		SpaceId:          spaceid,
		PackageVersionId: pkg.ActiveInstallID,
		InstalledId:      pkg.ID,
	}

	if err != nil {
		return nil, err
	}

	r.activeExecsLock.Lock()
	r.activeExecs[spaceid] = e
	r.activeExecsLock.Unlock()

	return e, nil

}

func (r *Runtime) ExecHttpQ(installedId, packageVersionId, spaceId int64, ctx *gin.Context) error {
	return r.ExecHttp(&xtypes.EngineHttpExecution{
		SpaceId:     spaceId,
		Request:     ctx,
		HandlerName: "",
		Params:      make(map[string]string),
	})
}

func (r *Runtime) ExecHttp(opts *xtypes.EngineHttpExecution) error {

	e, err := r.GetExec(opts.SpaceId)
	if err != nil {
		qq.Println("@exec_http/1", "error getting exec", err)
		httpx.WriteErr(opts.Request, err)
		return err
	}

	if e == nil {
		qq.Println("@exec_http/1", "exec is nil")
		httpx.WriteErr(opts.Request, errors.New("exec is nil"))
		return errors.New("exec is nil")
	}

	// print stack trace

	err = libx.PanicWrapper(func() {
		subpath := opts.Request.Param("subpath")

		opts.Params["space_id"] = fmt.Sprintf("%d", opts.SpaceId)
		opts.Params["install_id"] = fmt.Sprintf("%d", e.InstalledId)
		opts.Params["package_version_id"] = fmt.Sprintf("%d", e.PackageVersionId)
		opts.Params["subpath"] = subpath
		opts.Params["method"] = opts.Request.Request.Method

		if opts.HandlerName == "" {
			opts.HandlerName = "on_http"
		}

		err := e.Executor.HandleHttp(xtypes.HttpExecution{
			HandlerName: opts.HandlerName,
			Params:      opts.Params,
			Request:     opts.Request,
		})
		if err != nil {
			qq.Println("@exec_http/2", "error handling http", err)
			panic(err)
		}

	})

	return err

}

func (r *Runtime) ExecAction(opts *xtypes.EngineActionExecution) error {

	e, err := r.GetExec(opts.SpaceId)
	if err != nil {
		qq.Println("@exec_action/1", "error getting exec", err)
		return err
	}

	if e == nil {
		qq.Println("@exec_event/1", "exec is nil")
		return errors.New("exec is nil")
	}

	opts.Params["space_id"] = fmt.Sprintf("%d", opts.SpaceId)
	opts.Params["install_id"] = fmt.Sprintf("%d", e.InstalledId)
	opts.Params["package_version_id"] = fmt.Sprintf("%d", e.PackageVersionId)
	opts.Params["event_type"] = opts.ActionType
	opts.Params["action_name"] = opts.ActionName

	err = libx.PanicWrapper(func() {
		err := e.Executor.HandleAction(xtypes.ActionExecution{
			ActionType: opts.ActionType,
			ActionName: opts.ActionName,
			Params:     opts.Params,
			Request:    opts.Request,
		})
		if err != nil {
			qq.Println("@exec_event/2", "error handling event", err)
			panic(err)
		}
	})

	return err

}

func (r *Runtime) ExecActionQ(spaceId int64, req xtypes.ActionRequest, etype, actionName string) error {

	return r.ExecAction(&xtypes.EngineActionExecution{
		SpaceId:    spaceId,
		Request:    req,
		ActionType: etype,
		ActionName: actionName,
		Params:     make(map[string]string),
	})

}
