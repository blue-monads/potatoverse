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

func (b *RemoteHub) CorePublishEvent(ctx *HttpBindContext) (any, error) {
	opts := &core.PublishEventOptions{}
	err := ctx.Http.BindJSON(opts)
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
		InstallId:   ctx.PackageId,
		Name:        opts.Name,
		Payload:     payloadBytes,
		ResourceId:  opts.ResourceId,
		CollapseKey: opts.CollapseKey,
		SpaceId:     ctx.SpaceId,
	})
	return nil, err
}

func (b *RemoteHub) CoreFileToken(ctx *HttpBindContext) (any, error) {
	opts := &core.SignFsPresignedTokenOptions{}
	err := ctx.Http.BindJSON(opts)
	if err != nil {
		return nil, err
	}

	return b.signer.SignSpaceFilePresigned(&signer.SpaceFilePresignedClaim{
		InstallId: ctx.PackageId,
		UserId:    opts.UserId,
		PathName:  opts.Path,
		FileName:  opts.FileName,
	})
}

func (b *RemoteHub) CoreSignAdviseryToken(ctx *HttpBindContext) (any, error) {
	opts := &core.SignAdviseryTokenOptions{}
	err := ctx.Http.BindJSON(opts)
	if err != nil {
		return nil, err
	}

	return b.signer.SignSpaceAdvisiery(&signer.SpaceAdvisieryClaim{
		InstallId:    ctx.PackageId,
		UserId:       opts.UserId,
		TokenSubType: opts.SubType,
		Data:         opts.Data,
		SpaceId:      ctx.SpaceId,
	})
}

func (b *RemoteHub) CoreParseAdviseryToken(ctx *HttpBindContext) (any, error) {
	var req core.ParseTokenReq
	err := ctx.Http.BindJSON(&req)
	if err != nil {
		return nil, err
	}

	claim, err := b.signer.ParseSpaceAdvisiery(req.Token)
	if err != nil {
		return nil, err
	}

	if claim.InstallId != ctx.PackageId {
		return nil, errors.New("wrong install id")
	}

	if claim.SpaceId != ctx.SpaceId {
		return nil, errors.New("wrong space id")
	}

	return claim, nil
}

func (b *RemoteHub) CoreReadPackageFile(ctx *HttpBindContext) (any, error) {
	fpath := ctx.Http.Param("path")
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
	data, err := pops.GetFileContentByPath(ctx.PackageVersion, dirPath, fileName)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (b *RemoteHub) CoreListFiles(ctx *HttpBindContext) (any, error) {
	path := ctx.Http.Param("path")
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	return b.corehub.ListSpaceFilesSigned(ctx.PackageId, path)
}

func (b *RemoteHub) CoreDecodeFileId(ctx *HttpBindContext) (any, error) {
	id := ctx.Http.Param("id")
	return b.corehub.DecodeSpaceFileId(id)
}

func (b *RemoteHub) CoreEncodeFileId(ctx *HttpBindContext) (any, error) {
	idStr := ctx.Http.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, err
	}
	return b.corehub.EncodeSpaceFileId(id)
}

func (b *RemoteHub) CoreGetEnv(ctx *HttpBindContext) (any, error) {
	key := ctx.Http.Param("key")
	pkgOps := b.db.GetPackageInstallOps()
	pkg, err := pkgOps.GetPackage(ctx.PackageId)
	if err != nil {
		return nil, err
	}
	envs := make(map[string]string)
	if pkg.EnvVars != "" {
		if err := json.Unmarshal([]byte(pkg.EnvVars), &envs); err != nil {
			return nil, err
		}
	}
	return envs[key], nil
}
