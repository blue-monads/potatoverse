package server

import (
	"net/http"
	"time"

	"github.com/blue-monads/turnix/backend/app/actions"
	"github.com/blue-monads/turnix/backend/utils/libx/easyerr"
	"github.com/blue-monads/turnix/backend/utils/libx/httpx"
	"github.com/gin-gonic/gin"
)

func (a *Server) login(ctx *gin.Context) {
	data := &actions.LoginOpts{}

	err := ctx.Bind(data)
	if err != nil {
		httpx.WriteAuthErr(ctx, err)
		return
	}

	resp, err := a.ctrl.Login(data)
	if err != nil {
		httpx.WriteAuthErr(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)

}

func (a *Server) getInviteInfo(ctx *gin.Context) {
	token := ctx.Param("token")

	// Parse the invite token
	claim, err := a.signer.ParseInvite(token)
	if err != nil {
		httpx.WriteAuthErr(ctx, err)
		return
	}

	// Get invite details
	invite, err := a.ctrl.GetUserInvite(claim.InviteId)
	if err != nil {
		httpx.WriteAuthErr(ctx, err)
		return
	}

	// Check if invite is still valid
	if invite.Status != "pending" {
		httpx.WriteAuthErr(ctx, easyerr.Error("Invite has already been used"))
		return
	}

	if invite.ExpiresOn != nil && time.Now().After(*invite.ExpiresOn) {
		httpx.WriteAuthErr(ctx, easyerr.Error("Invite has expired"))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"email":      invite.Email,
		"role":       invite.InvitedAsType,
		"expires_on": invite.ExpiresOn,
	})
}

func (a *Server) acceptInvite(ctx *gin.Context) {
	token := ctx.Param("token")

	// Parse the invite token
	claim, err := a.signer.ParseInvite(token)
	if err != nil {
		httpx.WriteAuthErr(ctx, err)
		return
	}

	var req struct {
		Name     string `json:"name" binding:"required"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		httpx.WriteAuthErr(ctx, err)
		return
	}

	user, err := a.ctrl.AcceptUserInvite(claim.InviteId, req.Name, req.Username, req.Password)
	if err != nil {
		httpx.WriteAuthErr(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Account created successfully",
		"user":    user,
	})
}
