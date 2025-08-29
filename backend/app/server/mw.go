package server

import (
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/gin-gonic/gin"
)

type AuthedFunc func(claim *signer.AccessClaim, ctx *gin.Context) (any, error)

func (a *Server) withAccessTokenFn(fn AuthedFunc) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {

		claim, err := a.withAccessToken(ctx)
		if err != nil {
			return
		}

		resp, err := fn(claim, ctx)
		httpx.WriteJSON(ctx, resp, err)
	}

}

func (s *Server) withAccessToken(ctx *gin.Context) (*signer.AccessClaim, error) {

	tok := ctx.GetHeader("Authorization")
	claim, err := s.signer.ParseAccess(tok)
	if err != nil {
		httpx.WriteAuthErr(ctx, err)
		return nil, err
	}

	return claim, nil

}
