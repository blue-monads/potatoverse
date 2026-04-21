package remotehub

import (
	"crypto/sha256"
	"encoding/json"
	"errors"

	"golang.org/x/crypto/pbkdf2"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/caphub"
	"github.com/blue-monads/potatoverse/backend/services/corehub"
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	xutils "github.com/blue-monads/potatoverse/backend/utils"
	"github.com/blue-monads/potatoverse/backend/utils/libx/httpx"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
	"github.com/hako/branca"
)

type RemoteHub struct {
	tokenSigner *branca.Branca
	app         xtypes.App
	engine      xtypes.Engine
	caphub      *caphub.CapabilityHub
	corehub     *corehub.CoreHub
	signer      *signer.Signer
	db          datahub.Database
}

type HttpBindFn func(ctx *HttpBindContext) (any, error)

func NewRemoteHub() *RemoteHub {

	randomkey, err := xutils.GenerateRandomString(32)
	if err != nil {
		return nil
	}

	altKey := pbkdf2.Key([]byte(randomkey), []byte("EXE_UMAMI_POTATO"), 4, 32, sha256.New)

	return &RemoteHub{
		tokenSigner: branca.NewBranca(string(altKey)),
	}
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

func (b *RemoteHub) Init(app xtypes.App) {

	engine := app.Engine().(xtypes.Engine)
	corehub := app.CoreHub().(*corehub.CoreHub)
	signer := app.Signer()
	db := app.Database()
	caphub := engine.GetCapabilityHub().(*caphub.CapabilityHub)

	b.engine = engine
	b.corehub = corehub
	b.signer = signer
	b.db = db
	b.caphub = caphub

}

func (b *RemoteHub) Authed(h HttpBindFn) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		token := ctx.GetHeader(XExecHeader)
		if token == "" {
			httpx.WriteErr(ctx, errors.New("missing token"))
			return
		}

		payload, err := b.tokenSigner.DecodeToString(token)
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
			SpaceId:        claim.SpaceId,
			RequestId:      claim.RequestID,
		})

		httpx.WriteJSON(ctx, resp, err)
	}
}
