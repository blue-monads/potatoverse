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

		tok := ctx.GetHeader("Authorization")
		if tok == "" {
			httpx.WriteAuthErr(ctx, EmptyAuthTokenErr)
			return
		}

		claim, err := a.withAccessToken(tok)
		if err != nil {
			return
		}

		resp, err := fn(claim, ctx)
		if resp == nil && err == nil {
			return
		}

		httpx.WriteJSON(ctx, resp, err)
	}

}

var EmptyAuthTokenErr = errors.New("empty auth token")

func (s *Server) withAccessToken(tok string) (*signer.AccessClaim, error) {

	finalTok := strings.TrimPrefix(tok, "TokenV1 ")

	claim, err := s.signer.ParseAccess(finalTok)
	if err != nil {
		return nil, err
	}

	return claim, nil

}
