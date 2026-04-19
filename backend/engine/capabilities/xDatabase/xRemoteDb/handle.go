package xremotedb

import (
	"github.com/blue-monads/potatoverse/backend/utils/libx/httpx"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/gin-gonic/gin"
)

func (r *RemoteDbCapability) Handle(ctx *gin.Context) {
	claim, err := r.capHandle.ValidateCapToken(ctx.Request.Header.Get("Authorization"))
	if err != nil {
		httpx.WriteErrString(ctx, "invalid token")
		return
	}

	model := r.capHandle.GetModel()
	if model.ID != claim.CapabilityId {
		httpx.WriteErrString(ctx, "invalid capability id")
		return
	}

	action := ctx.Param("subpath")
	if action == "" {
		httpx.WriteErrString(ctx, "action is required")
		return
	}

	result, err := r.Execute(action, lazydata.NewLazyHTTP(ctx))
	httpx.WriteJSON(ctx, result, err)

}
