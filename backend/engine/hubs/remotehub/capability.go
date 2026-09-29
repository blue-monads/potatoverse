package remotehub

import (
	"github.com/blue-monads/potatoverse/backend/engine/executors/core"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
)

func (b *RemoteHub) CapTokenSign(ctx RContext) (any, error) {
	capName, err := ctx.GetMeta("cap")
	if err != nil {
		return nil, err
	}
	var opts core.CapTokenSignOptions
	err = bindJSON(ctx, &opts)
	if err != nil {
		return nil, err
	}

	capability, err := b.db.GetSpaceOps().GetSpaceCapability(ctx.GetPackageId(), capName)
	if err != nil {
		return nil, err
	}

	res, err := b.signer.SignCapability(&signer.CapabilityClaim{
		CapabilityId: capability.ID,
		InstallId:    ctx.GetPackageId(),
		SpaceId:      ctx.GetSpaceId(),
		UserId:       opts.UserId,
		ResourceId:   opts.ResourceId,
		SubType:      opts.SubType,
		ExtraMeta:    opts.ExtraMeta,
	})
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) CapList(ctx RContext) (any, error) {
	res, err := b.caphub.List(ctx.GetSpaceId())
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) CapExecute(ctx RContext) (any, error) {
	method, err := ctx.GetMeta("method")
	if err != nil {
		return nil, err
	}
	capName, err := ctx.GetMeta("cap")
	if err != nil {
		return nil, err
	}

	data, err := ctx.GetData()
	if err != nil {
		return nil, err
	}

	var lh lazydata.LazyData = lazydata.LazyDataBytes(data)

	res, err := b.caphub.Execute(ctx.GetPackageId(), ctx.GetSpaceId(), capName, method, lh)
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) CapMethods(ctx RContext) (any, error) {
	capName, err := ctx.GetMeta("cap")
	if err != nil {
		return nil, err
	}
	res, err := b.caphub.Methods(ctx.GetPackageId(), ctx.GetSpaceId(), capName)
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}
