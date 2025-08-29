package server

import (
	"errors"
	"strings"

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

var EmptyAuthTokenErr = errors.New("empty auth token")

func (s *Server) withAccessToken(ctx *gin.Context) (*signer.AccessClaim, error) {

	tok := ctx.GetHeader("Authorization")
	if tok == "" {
		httpx.WriteAuthErr(ctx, EmptyAuthTokenErr)
		return nil, EmptyAuthTokenErr
	}

	finalTok := strings.TrimPrefix(tok, "TokenV1 ")

	claim, err := s.signer.ParseAccess(finalTok)
	if err != nil {
		httpx.WriteAuthErr(ctx, err)
		return nil, err
	}

	return claim, nil

}
