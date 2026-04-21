package remotehub

import (
	"github.com/blue-monads/potatoverse/backend/services/signer"
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

func (b *RemoteHub) CapTokenSign(ctx *HttpBindContext) (any, error) {
	capName := ctx.Http.Param("cap")
	var opts struct {
		ResourceId string         `json:"resource_id"`
		ExtraMeta  map[string]any `json:"extrameta"`
		UserId     int64          `json:"user_id"`
		SubType    string         `json:"sub_type"`
	}
	err := ctx.Http.BindJSON(&opts)
	if err != nil {
		return nil, err
	}

	capability, err := b.db.GetSpaceOps().GetSpaceCapability(ctx.PackageId, capName)
	if err != nil {
		return nil, err
	}

	return b.signer.SignCapability(&signer.CapabilityClaim{
		CapabilityId: capability.ID,
		InstallId:    ctx.PackageId,
		SpaceId:      ctx.SpaceId,
		UserId:       opts.UserId,
		ResourceId:   opts.ResourceId,
		SubType:      opts.SubType,
		ExtraMeta:    opts.ExtraMeta,
	})
}

func (b *RemoteHub) CapList(ctx *HttpBindContext) (any, error) {
	return b.caphub.List(ctx.SpaceId)
}

func (b *RemoteHub) CapExecute(ctx *HttpBindContext) (any, error) {
	method := ctx.Http.Param("method")
	capName := ctx.Http.Param("cap")

	lh := lazydata.NewLazyHTTP(ctx.Http)

	return b.caphub.Execute(ctx.PackageId, ctx.SpaceId, capName, method, lh)
}

func (b *RemoteHub) CapMethods(ctx *HttpBindContext) (any, error) {
	capName := ctx.Http.Param("cap")
	return b.caphub.Methods(ctx.PackageId, ctx.SpaceId, capName)
}
