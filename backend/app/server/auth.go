package server

import (
	"net/http"

	"github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/gin-gonic/gin"
)

func (a *Server) login(ctx *gin.Context) {
	data := &actions.LoginOpts{}

	// LoginOpts
	err := ctx.Bind(data)
	if err != nil {
		httpx.WriteAuthErr(ctx, err)
		return
	}

	token, err := a.ctrl.Login(*data)
	if err != nil {
		httpx.WriteAuthErr(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token": token,
	})

}
