package rtbinds

import (
	"crypto/sha256"
	"encoding/json"
	"errors"

	"golang.org/x/crypto/pbkdf2"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/caphub"
	"github.com/blue-monads/potatoverse/backend/services/corehub"
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/utils/libx/httpx"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
	"github.com/hako/branca"
)

type BindServer struct {
	tokenSigner *branca.Branca
	app         xtypes.App
	engine      xtypes.Engine
	caphub      *caphub.CapabilityHub
	corehub     *corehub.CoreHub
	signer      *signer.Signer
	db          datahub.Database
}

type HttpBindFn func(ctx *HttpBindContext) (any, error)

func NewBindServer(app xtypes.App, key []byte) *BindServer {
	altKey := pbkdf2.Key(key, []byte("EXE_UMAMI_POTATO"), 4, 32, sha256.New)

	engine := app.Engine().(xtypes.Engine)
	corehub := app.CoreHub().(*corehub.CoreHub)
	signer := app.Signer()
	db := app.Database()
	caphub := engine.GetCapabilityHub().(*caphub.CapabilityHub)

	return &BindServer{
		tokenSigner: branca.NewBranca(string(altKey)),
		app:         app,
		engine:      engine,
		corehub:     corehub,
		signer:      signer,
		db:          db,
		caphub:      caphub,
	}
}

func (b *BindServer) BindEngine(rg *gin.RouterGroup) {

	rg.GET("/capability/list", b.authed(b.CapList))
	rg.POST("/capability/:cap/sign", b.authed(b.CapTokenSign))
	rg.GET("/capability/:cap/methods", b.authed(b.CapMethods))
	rg.POST("/capability/:cap/execute/:method", b.authed(b.CapExecute))

	rg.GET("/core/read_package_file/*path", b.authed(b.CoreReadPackageFile))
	rg.GET("/core/list_files/*path", b.authed(b.CoreListFiles))
	rg.GET("/core/decode_file_id/:id", b.authed(b.CoreDecodeFileId))
	rg.GET("/core/encode_file_id/:id", b.authed(b.CoreEncodeFileId))
	rg.GET("/core/env/:key", b.authed(b.CoreGetEnv))

	rg.POST("/core/publish_event", b.authed(b.CorePublishEvent))
	rg.POST("/core/file_token", b.authed(b.CoreFileToken))
	rg.POST("/core/sign_advisery_token", b.authed(b.CoreSignAdviseryToken))
	rg.POST("/core/parse_advisery_token", b.authed(b.CoreParseAdviseryToken))

	rg.POST("/db/run_query", b.authed(b.DBRunQuery))
	rg.POST("/db/run_query_one", b.authed(b.DBRunQueryOne))
	rg.POST("/db/insert", b.authed(b.DBInsert))
	rg.POST("/db/update_by_id", b.authed(b.DBUpdateById))
	rg.POST("/db/delete_by_id", b.authed(b.DBDeleteById))
	rg.POST("/db/find_by_id", b.authed(b.DBFindById))
	rg.POST("/db/update_by_cond", b.authed(b.DBUpdateByCond))
	rg.POST("/db/delete_by_cond", b.authed(b.DBDeleteByCond))
	rg.POST("/db/find_all_by_cond", b.authed(b.DBFindAllByCond))
	rg.POST("/db/find_one_by_cond", b.authed(b.DBFindOneByCond))
	rg.POST("/db/find_all_by_query", b.authed(b.DBFindAllByQuery))
	rg.POST("/db/find_by_join", b.authed(b.DBFindByJoin))
	rg.GET("/db/list_tables", b.authed(b.DBListTables))
	rg.GET("/db/table/:table/columns", b.authed(b.DBListColumns))

	rg.POST("/kv/add", b.authed(b.KVAdd))
	rg.GET("/kv/:group/:key", b.authed(b.KVGet))
	rg.POST("/kv/query", b.authed(b.KVQuery))
	rg.POST("/kv/remove", b.authed(b.KVRemove))
	rg.POST("/kv/update", b.authed(b.KVUpdate))
	rg.POST("/kv/upsert", b.authed(b.KVUpsert))

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
