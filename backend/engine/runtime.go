package engine

import (
	"sync"
	"time"

	"github.com/blue-monads/turnix/backend/engine/luaz"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type Runtime struct {
	execs     map[int64]*luaz.Luaz
	execsLock sync.RWMutex
	parent    *Engine
}

func (r *Runtime) cleanupExecs() {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for range ticker.C {
		r.execsLock.RLock()
		for _, e := range r.execs {
			e.Cleanup()
		}

		r.execsLock.RUnlock()
	}
}

func (r *Runtime) GetExec(packageId int64) *luaz.Luaz {
	r.execsLock.RLock()
	e := r.execs[packageId]
	r.execsLock.RUnlock()
	if e != nil {
		return e
	}

	file, err := r.parent.db.GetPackageFileMetaByPath(packageId, "", "main.lua")
	if err != nil {
		r.parent.app.Logger().Error("error getting package file meta by path", "error", err)
		return nil
	}

	sourceBytes, err := r.parent.db.GetPackageFile(packageId, file.ID)
	if err != nil {
		r.parent.app.Logger().Error("error getting package file", "error", err)
		return nil
	}

	e = luaz.New(luaz.Options{
		BuilderOpts: xtypes.BuilderOption{
			App:    r.parent.app,
			Logger: r.parent.app.Logger().With("package_id", packageId),
		},
		Code:          string(sourceBytes),
		WorkingFolder: "",
	})

	r.execsLock.Lock()
	r.execs[packageId] = e
	r.execsLock.Unlock()

	return e

}

func (r *Runtime) ExecHttp(packageId, spaceId int64, ctx *gin.Context) {

	e := r.GetExec(packageId)
	if e == nil {
		return
	}

	e.Handle(ctx, "on_http")

}
