package server

import (
	"github.com/blue-monads/potatoverse/backend/services/signer"
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

func (s *Server) selfInfo(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	user, err := s.ctrl.GetSelfInfo(claim.UserId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

type UpdateBioRequest struct {
	Bio string `json:"bio"`
}

func (s *Server) updateSelfBio(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	var req UpdateBioRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	err := s.ctrl.UpdateSelfBio(claim.UserId, req.Bio)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Bio updated successfully"}, nil
}
