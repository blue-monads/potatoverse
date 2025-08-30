package server

import (
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/gin-gonic/gin"
)

type BPrint struct {
	Name            string `json:"name"`
	RunTimeVersion  string `json:"runtime_version"`
	Executor        string `json:"executor"`
	ExecutorVersion string `json:"executor_version"`
	HomePage        string `json:"home_page"`
	Logo            string `json:"logo"`
}

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
