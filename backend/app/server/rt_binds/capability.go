package rtbinds

import (
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/gin-gonic/gin"
)

type HttpBindContext struct {
	Http           *gin.Context
	PackageId      int64
	PackageVersion int64
	SpaceId        int64
	RequestId      string
}

func (b *BindServer) CapTokenSign(ctx *HttpBindContext) (any, error) {

	return nil, nil
}

func (b *BindServer) CapList(ctx *HttpBindContext) (any, error) {

	return nil, nil
}

func (b *BindServer) CapExecute(ctx *HttpBindContext) (any, error) {
	method := ctx.Http.Query("method")
	capName := ctx.Http.Query("cap")

	lh := lazydata.NewLazyHTTP(ctx.Http)

	return b.caphub.Execute(ctx.PackageId, ctx.SpaceId, capName, method, lh)
}

func (b *BindServer) CapMethods(ctx *HttpBindContext) (any, error) {
	capName := ctx.Http.Query("cap")
	return b.caphub.Methods(ctx.PackageId, ctx.SpaceId, capName)
}
