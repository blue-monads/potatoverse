package server

import (
	"strconv"

	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/gin-gonic/gin"
)

func (s *Server) listUsers(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {

	offset, _ := strconv.Atoi(ctx.Query("offset"))
	limit, _ := strconv.Atoi(ctx.Query("limit"))

	if limit == 0 {
		limit = 100
	}

	users, err := s.ctrl.ListUsers(offset, limit)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *Server) getUser(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {

	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	user, err := s.ctrl.GetUser(int64(idInt))
	if err != nil {
		return nil, err
	}

	return user, nil
}

// User Invite handlers

func (s *Server) listUserInvites(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	offset, _ := strconv.Atoi(ctx.Query("offset"))
	limit, _ := strconv.Atoi(ctx.Query("limit"))

	if limit == 0 {
		limit = 100
	}

	invites, err := s.ctrl.ListUserInvites(offset, limit)
	if err != nil {
		return nil, err
	}

	return invites, nil
}

func (s *Server) getUserInvite(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	invite, err := s.ctrl.GetUserInvite(int64(idInt))
	if err != nil {
		return nil, err
	}

	return invite, nil
}

func (s *Server) addUserInvite(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	var req struct {
		Email         string `json:"email" binding:"required,email"`
		InvitedAsType string `json:"invited_as_type" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	// fixme => future
	role := "normal"

	invite, err := s.ctrl.AddUserInvite(req.Email, role, req.InvitedAsType, claim.UserId)
	if err != nil {
		return nil, err
	}

	return invite, nil
}

func (s *Server) updateUserInvite(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	var req map[string]any
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	err = s.ctrl.UpdateUserInvite(int64(idInt), req)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "User invite updated successfully"}, nil
}

func (s *Server) deleteUserInvite(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	err = s.ctrl.DeleteUserInvite(int64(idInt))
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "User invite deleted successfully"}, nil
}

func (s *Server) resendUserInvite(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	invite, err := s.ctrl.ResendUserInvite(int64(idInt))
	if err != nil {
		return nil, err
	}

	return invite, nil
}

// Create User Directly

func (s *Server) createUserDirectly(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Username string `json:"username" binding:"required"`
		Utype    string `json:"utype" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	user, err := s.ctrl.CreateUserDirectly(req.Name, req.Email, req.Username, req.Utype, claim.UserId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

/*
AddUser
ResetUserPassword
DeactivateUser
ActivateUser
DeleteUser
UpdateUser
*/
