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
		qq.Println("@get_exec/1", "error getting space", err)
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
	return r.ExecHttp(&xtypes.HttpEventOptions{
		SpaceId:     spaceId,
		Request:     ctx,
		HandlerName: "",
		Params:      make(map[string]string),
	})
}

func (r *Runtime) ExecHttp(opts *xtypes.HttpEventOptions) error {

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

	subpath := opts.Request.Param("subpath")

	if opts.Params == nil {
		opts.Params = make(map[string]string)
	}

	opts.Params["space_id"] = fmt.Sprintf("%d", opts.SpaceId)
	opts.Params["install_id"] = fmt.Sprintf("%d", e.InstalledId)
	opts.Params["package_version_id"] = fmt.Sprintf("%d", e.PackageVersionId)
	opts.Params["subpath"] = subpath
	opts.Params["method"] = opts.Request.Request.Method

	// print stack trace

	err = libx.PanicWrapper(func() {

		if opts.HandlerName == "" {
			opts.HandlerName = "on_http"
		}

		qq.Println("@exec_http/3", opts.Params)

		err := e.Executor.HandleHttp(&xtypes.HttpEvent{
			EventType:   opts.EventType,
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

func (r *Runtime) ExecAction(opts *xtypes.ActionEventOptions) error {

	e, err := r.GetExec(opts.SpaceId)
	if err != nil {
		qq.Println("@exec_action/1", "error getting exec", err)
		return err
	}

	if e == nil {
		qq.Println("@exec_event/1", "exec is nil")
		return errors.New("exec is nil")
	}

	if opts.Params == nil {
		opts.Params = make(map[string]string)
	}

	opts.Params["space_id"] = fmt.Sprintf("%d", opts.SpaceId)
	opts.Params["install_id"] = fmt.Sprintf("%d", e.InstalledId)
	opts.Params["package_version_id"] = fmt.Sprintf("%d", e.PackageVersionId)
	opts.Params["event_type"] = opts.EventType
	opts.Params["action"] = opts.ActionName

	err = libx.PanicWrapper(func() {

		qq.Println("@params", opts.Params)

		err := e.Executor.HandleAction(&xtypes.ActionEvent{
			EventType:  opts.EventType,
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

	return r.ExecAction(&xtypes.ActionEventOptions{
		SpaceId:    spaceId,
		Request:    req,
		EventType:  etype,
		ActionName: actionName,
		Params:     make(map[string]string),
	})

}
