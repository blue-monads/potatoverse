package server

import (
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/gin-gonic/gin"
)

func (s *Server) Doc(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	return gin.H{"message": "Doc"}, nil
}
