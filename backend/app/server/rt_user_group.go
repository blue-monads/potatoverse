package server

import (
	"errors"

	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/gin-gonic/gin"
)

// User Group handlers

func (s *Server) listUserGroups(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	groups, err := s.ctrl.ListUserGroups()
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (s *Server) getUserGroup(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	name := ctx.Param("name")
	if name == "" {
		return nil, errors.New("name parameter is required")
	}

	group, err := s.ctrl.GetUserGroup(name)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s *Server) addUserGroup(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Info string `json:"info"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	err := s.ctrl.AddUserGroup(req.Name, req.Info)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "User group created successfully"}, nil
}

func (s *Server) updateUserGroup(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	name := ctx.Param("name")
	if name == "" {
		return nil, errors.New("name parameter is required")
	}

	var req struct {
		Info string `json:"info"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	err := s.ctrl.UpdateUserGroup(name, req.Info)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "User group updated successfully"}, nil
}

func (s *Server) deleteUserGroup(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	name := ctx.Param("name")
	if name == "" {
		return nil, errors.New("name parameter is required")
	}

	err := s.ctrl.DeleteUserGroup(name)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "User group deleted successfully"}, nil
}
