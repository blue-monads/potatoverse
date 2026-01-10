package server

import (
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/utils/libx/httpx"
	"github.com/gin-gonic/gin"
)

func (s *Server) parseSpaceToken(ctx *gin.Context) (*signer.SpaceClaim, error) {
	tok := ctx.GetHeader("Autorization")

	// fixme =>  check permission

	claim, err := s.signer.ParseSpace(tok)
	if err != nil {
		return nil, err
	}

	return claim, nil
}

func (s *Server) signedFileList(ctx *gin.Context) {
	claim, err := s.parseSpaceToken(ctx)
	if err != nil {
		httpx.UnAuthorized(ctx)
		return
	}

	path := ctx.Query("path")

	files, err := s.opt.CoreHub.ListSpaceFilesSigned(claim.InstallId, path)
	if err != nil {
		httpx.WriteErr(ctx, err)
		return
	}

	httpx.WriteJSON(ctx, files, nil)
}

func (s *Server) signedFileDownload(ctx *gin.Context) {

	refId := ctx.Param("ref_id")

	s.opt.CoreHub.ServeSpaceFileSigned(refId, ctx)
}

func (s *Server) signedFilePreview(ctx *gin.Context) {
	refId := ctx.Param("ref_id")

	s.opt.CoreHub.ServePreviewFileSigned(refId, ctx)
}

func (s *Server) signedFileUpload(ctx *gin.Context)       {}
func (s *Server) signedFileCreateFolder(ctx *gin.Context) {}
