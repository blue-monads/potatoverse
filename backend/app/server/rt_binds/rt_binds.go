package rtbinds

import (
	"crypto/sha256"
	"encoding/json"
	"errors"

	"golang.org/x/crypto/pbkdf2"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/caphub"
	"github.com/blue-monads/potatoverse/backend/utils/libx/httpx"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
	"github.com/hako/branca"
)

type BindServer struct {
	signer *branca.Branca
	engine xtypes.Engine

	caphub *caphub.CapabilityHub
}

type HttpBindFn func(ctx *HttpBindContext) (any, error)

func NewBindServer(key []byte) *BindServer {
	altKey := pbkdf2.Key(key, []byte("EXE_UMAMI_POTATO"), 4, 32, sha256.New)
	return &BindServer{
		signer: branca.NewBranca(string(altKey)),
	}
}

func (b *BindServer) BindEngine(rg *gin.RouterGroup) {

	rg.GET("/capability/sign", b.authed(b.CapTokenSign))
	rg.GET("/capability/methods", b.authed(b.CapMethods))
	rg.GET("/capability/list", b.authed(b.CapList))
	rg.GET("/capability/execute", b.authed(b.CapExecute))

}

const (
	XExecHeader = "X-Exec-Header"
)

type XExecClaim struct {
	PackageId        int64  `json:"p"`
	PackageVersionId int64  `json:"v"`
	SpaceId          int64  `json:"s"`
	RequestID        string `json:"r"`
}

func (b *BindServer) authed(h HttpBindFn) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		token := ctx.GetHeader(XExecHeader)
		if token == "" {
			httpx.WriteErr(ctx, errors.New("missing token"))
			return
		}

		payload, err := b.signer.DecodeToString(token)
		if err != nil {
			httpx.WriteErr(ctx, errors.New("invalid token"))
			return
		}

		claim := &XExecClaim{}

		err = json.Unmarshal([]byte(payload), claim)

		if err != nil {
			httpx.WriteErr(ctx, errors.New("invalid token"))
			return
		}

		resp, err := h(&HttpBindContext{
			Http:           ctx,
			PackageId:      claim.PackageId,
			PackageVersion: claim.PackageVersionId,
		})

		httpx.WriteJSON(ctx, resp, err)
	}
}
