package server

import (
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/gin-gonic/gin"
)

// selfUserPortalData

type AdminPortalData struct {
	PopularKeywords  []string `json:"popular_keywords"`
	FavoriteProjects []any    `json:"favorite_projects"`
}

func (s *Server) selfUserPortalData(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {

	return AdminPortalData{
		PopularKeywords: []string{
			"game",
			"agent",
		},
		FavoriteProjects: []any{},
	}, nil

}
