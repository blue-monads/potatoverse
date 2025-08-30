package server

import (
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/gin-gonic/gin"
)

// selfUserPortalData
func (s *Server) selfUserPortalData(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {

	return map[string]any{
		"popular_keywords": []string{
			"game",
			"agent",
		},
		"featured_projects": []any{},
	}, nil

}
