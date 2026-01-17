package server

import (
	"strconv"

	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/gin-gonic/gin"
)

func (a *Server) GetInstalledPackageInfo(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	packageId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		return nil, err
	}

	packageInfo, err := a.ctrl.GetInstalledPackageInfo(packageId)
	if err != nil {
		return nil, err
	}

	return packageInfo, nil
}
