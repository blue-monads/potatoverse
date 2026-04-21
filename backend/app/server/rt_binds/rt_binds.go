package rtbinds

import (
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/caphub"
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
	altKey := pbkdf2.Key(key, []byte("UMAMI_POTATO"), 4, 32, sha256.New)
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

func (b *BindServer) authed(h HttpBindFn) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h(&HttpBindContext{
			Http:           ctx,
			PackageId:      0,
			PackageVersion: 0,
		})
	}
}
