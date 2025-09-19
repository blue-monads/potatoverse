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

/*

AddUser
ResetUserPassword
DeactivateUser
ActivateUser
DeleteUser
UpdateUser



*/
