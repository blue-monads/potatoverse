package rtbinds

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

type PublishEventOptions struct {
	Name        string `json:"name"`
	Payload     any    `json:"payload"`
	ResourceId  string `json:"resource_id"`
	CollapseKey string `json:"collapse_key"`
}

func (b *BindServer) CorePublishEvent(ctx *HttpBindContext) (any, error) {
	opts := &PublishEventOptions{}
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

type SignFsPresignedTokenOptions struct {
	Path     string `json:"path"`
	FileName string `json:"file_name"`
	UserId   int64  `json:"user_id"`
}

func (b *BindServer) CoreFileToken(ctx *HttpBindContext) (any, error) {
	opts := &SignFsPresignedTokenOptions{}
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

type SignAdviseryTokenOptions struct {
	TokenSubType string         `json:"token_sub_type"`
	UserId       int64          `json:"user_id"`
	Data         map[string]any `json:"data"`
}

func (b *BindServer) CoreSignAdviseryToken(ctx *HttpBindContext) (any, error) {
	opts := &SignAdviseryTokenOptions{}
	err := ctx.Http.BindJSON(opts)
	if err != nil {
		return nil, err
	}

	return b.signer.SignSpaceAdvisiery(&signer.SpaceAdvisieryClaim{
		InstallId:    ctx.PackageId,
		UserId:       opts.UserId,
		TokenSubType: opts.TokenSubType,
		Data:         opts.Data,
		SpaceId:      ctx.SpaceId,
	})
}

func (b *BindServer) CoreParseAdviseryToken(ctx *HttpBindContext) (any, error) {
	var req struct {
		Token string `json:"token"`
	}
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

func (b *BindServer) CoreReadPackageFile(ctx *HttpBindContext) (any, error) {
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

func (b *BindServer) CoreListFiles(ctx *HttpBindContext) (any, error) {
	path := ctx.Http.Param("path")
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	return b.corehub.ListSpaceFilesSigned(ctx.PackageId, path)
}

func (b *BindServer) CoreDecodeFileId(ctx *HttpBindContext) (any, error) {
	id := ctx.Http.Param("id")
	return b.corehub.DecodeSpaceFileId(id)
}

func (b *BindServer) CoreEncodeFileId(ctx *HttpBindContext) (any, error) {
	idStr := ctx.Http.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, err
	}
	return b.corehub.EncodeSpaceFileId(id)
}

func (b *BindServer) CoreGetEnv(ctx *HttpBindContext) (any, error) {
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
