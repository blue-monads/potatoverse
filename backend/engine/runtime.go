package engine

import (
	"sync"

	"github.com/blue-monads/turnix/backend/engine/luaz"
	"github.com/gin-gonic/gin"
)

type Runtime struct {
	execs     map[int64]*luaz.Luaz
	execsLock sync.RWMutex
}

func (r *Runtime) GetExec(packageId int64) *luaz.Luaz {
	r.execsLock.RLock()
	defer r.execsLock.RUnlock()
	e := r.execs[packageId]
	if e == nil {
		e = luaz.New(luaz.Options{
			// todo => builder opts
			Code:          "",
			WorkingFolder: "",
		})
	}
	return e
}

func (r *Runtime) ExecHttp(packageId, spaceId int64, ctx *gin.Context) {

	e := r.GetExec(packageId)
	if e == nil {
		return
	}

	e.Handle(ctx, "on_http")

}
