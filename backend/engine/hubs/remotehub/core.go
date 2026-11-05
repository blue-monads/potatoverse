package remotehub

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/blue-monads/potatoverse/backend/engine/executors/core"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

func (b *RemoteHub) CorePublishEvent(ctx RContext) (any, error) {
	opts := &core.PublishEventOptions{}
	err := bindJSON(ctx, opts)
	if err != nil {
		return nil, err
	}

	var payloadBytes []byte
	if opts.Payload == nil {
		payloadBytes = []byte{}
	} else {
		switch v := opts.Payload.(type) {
		case string:
			payloadBytes = []byte(v)
		case []byte:
			payloadBytes = v
		default:
			jsonData, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			payloadBytes = jsonData
		}
	}

	err = b.engine.PublishEvent(&xtypes.EventOptions{
		InstallId:   ctx.GetPackageId(),
		Name:        opts.Name,
		Payload:     payloadBytes,
		ResourceId:  opts.ResourceId,
		CollapseKey: opts.CollapseKey,
		SpaceId:     ctx.GetSpaceId(),
	})
	return nil, err
}

func (b *RemoteHub) CoreFileToken(ctx RContext) (any, error) {
	opts := &core.SignFsPresignedTokenOptions{}
	err := bindJSON(ctx, opts)
	if err != nil {
		return nil, err
	}

	res, err := b.signer.SignSpaceFilePresigned(&signer.SpaceFilePresignedClaim{
		InstallId: ctx.GetPackageId(),
		UserId:    opts.UserId,
		PathName:  opts.Path,
		FileName:  opts.FileName,
	})
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) CoreSignAdviseryToken(ctx RContext) (any, error) {
	opts := &core.SignAdviseryTokenOptions{}
	err := bindJSON(ctx, opts)
	if err != nil {
		return nil, err
	}

	res, err := b.signer.SignSpaceAdvisiery(&signer.SpaceAdvisieryClaim{
		InstallId:    ctx.GetPackageId(),
		UserId:       opts.UserId,
		TokenSubType: opts.SubType,
		Data:         opts.Data,
		SpaceId:      ctx.GetSpaceId(),
	})
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) CoreParseAdviseryToken(ctx RContext) (any, error) {
	var req core.ParseTokenReq
	err := bindJSON(ctx, &req)
	if err != nil {
		return nil, err
	}

	claim, err := b.signer.ParseSpaceAdvisiery(req.Token)
	if err != nil {
		return nil, err
	}

	if claim.InstallId != ctx.GetPackageId() {
		return nil, errors.New("wrong install id")
	}

	if claim.SpaceId != ctx.GetSpaceId() {
		return nil, errors.New("wrong space id")
	}

	setDataJSON(ctx, claim)
	return claim, nil
}

func (b *RemoteHub) CoreReadPackageFile(ctx RContext) (any, error) {
	fpath, err := ctx.GetMeta("path")
	if err != nil {
		return nil, err
	}
	if len(fpath) > 0 && fpath[0] == '/' {
		fpath = fpath[1:]
	}
	fileName := fpath
	dirPath := ""

	if strings.Contains(fpath, "/") {
		parts := strings.Split(fpath, "/")
		fileName = parts[len(parts)-1]
		dirPath = strings.Join(parts[:len(parts)-1], "/")
	}

	pops := b.db.GetPackageFileOps()
	data, err := pops.GetFileContentByPath(ctx.GetPackageVersion(), dirPath, fileName)
	if err != nil {
		return nil, err
	}
	_ = ctx.SetData(data)
	return string(data), nil
}

func (b *RemoteHub) CoreListFiles(ctx RContext) (any, error) {
	path, _ := ctx.GetMeta("path")
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	res, err := b.corehub.ListSpaceFilesSigned(ctx.GetPackageId(), path)
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) CoreDecodeFileId(ctx RContext) (any, error) {
	id, err := ctx.GetMeta("id")
	if err != nil {
		return nil, err
	}
	res, err := b.corehub.DecodeSpaceFileId(id)
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) CoreEncodeFileId(ctx RContext) (any, error) {
	idStr, err := ctx.GetMeta("id")
	if err != nil {
		return nil, err
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, err
	}
	res, err := b.corehub.EncodeSpaceFileId(id)
	if err != nil {
		return nil, err
	}
	setDataJSON(ctx, res)
	return res, nil
}

func (b *RemoteHub) CoreGetEnv(ctx RContext) (any, error) {
	key, err := ctx.GetMeta("key")
	if err != nil {
		return nil, err
	}
	pkgOps := b.db.GetPackageInstallOps()
	pkg, err := pkgOps.GetPackage(ctx.GetPackageId())
	if err != nil {
		return nil, err
	}
	envs := make(map[string]string)
	if pkg.EnvVars != "" {
		if err := json.Unmarshal([]byte(pkg.EnvVars), &envs); err != nil {
			return nil, err
		}
	}
	res := envs[key]
	_ = ctx.SetData([]byte(res))
	return res, nil
}
