package server

import (
	"strconv"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/utils/libx/httpx"
	"github.com/gin-gonic/gin"
)

func (s *Server) sendUserMessage(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	var req struct {
		Title         string `json:"title" binding:"required"`
		Type          string `json:"type" binding:"required"`
		Contents      string `json:"contents" binding:"required"`
		ToUser        int64  `json:"to_user" binding:"required"`
		FromSpaceId   int64  `json:"from_space_id"`
		CallbackToken string `json:"callback_token"`
		WarnLevel     int    `json:"warn_level"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	msg := &dbmodels.UserMessage{
		Title:         req.Title,
		Type:          req.Type,
		Contents:      req.Contents,
		ToUser:        req.ToUser,
		FromUserId:    claim.UserId,
		FromSpaceId:   req.FromSpaceId,
		CallbackToken: req.CallbackToken,
		WarnLevel:     req.WarnLevel,
		IsRead:        false,
	}

	coreHub := s.opt.CoreHub

	id, err := coreHub.UserSendMessage(msg)
	if err != nil {
		return nil, err
	}

	msg.ID = id
	return msg, nil
}

func (s *Server) queryNewMessages(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	// Get user's read head
	readHead, err := s.ctrl.GetUserReadHead(claim.UserId)
	if err != nil {
		return nil, err
	}

	messages, err := s.ctrl.QueryNewMessages(claim.UserId, readHead)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *Server) queryMessageHistory(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	limitStr := ctx.Query("limit")
	limit := 100
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 1000 {
			limit = 100
		}
	}

	messages, err := s.ctrl.QueryMessageHistory(claim.UserId, limit)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *Server) setAllMessagesAsRead(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	// Get the highest message ID for this user
	messages, err := s.ctrl.ListUserMessages(claim.UserId, 0, 1)
	if err != nil {
		return nil, err
	}

	var readHead int64 = 0
	if len(messages) > 0 {
		readHead = messages[0].ID
	}

	err = s.ctrl.SetAllMessagesAsRead(claim.UserId, readHead)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "All messages marked as read", "read_head": readHead}, nil
}

func (s *Server) setMessageAsRead(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	err = s.ctrl.SetMessageAsRead(int64(idInt), claim.UserId)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Message marked as read"}, nil
}

func (s *Server) listUserMessages(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	afterIdStr := ctx.Query("after_id")
	limit, _ := strconv.Atoi(ctx.Query("limit"))

	if limit == 0 {
		limit = 100
	}

	var afterId int64 = 0
	if afterIdStr != "" {
		id, err := strconv.ParseInt(afterIdStr, 10, 64)
		if err == nil {
			afterId = id
		}
	}

	messages, err := s.ctrl.ListUserMessages(claim.UserId, afterId, limit)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *Server) getUserMessage(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	message, err := s.ctrl.GetUserMessage(int64(idInt))
	if err != nil {
		return nil, err
	}

	// Verify the message belongs to the user
	if message.ToUser != claim.UserId {
		return nil, gin.Error{Err: gin.Error{Err: nil}, Type: gin.ErrorTypePrivate, Meta: "unauthorized"}
	}

	return message, nil
}

func (s *Server) updateUserMessage(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	var req map[string]any
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	// Verify the message belongs to the user before updating
	message, err := s.ctrl.GetUserMessage(int64(idInt))
	if err != nil {
		return nil, err
	}
	if message.ToUser != claim.UserId {
		return nil, gin.Error{Err: gin.Error{Err: nil}, Type: gin.ErrorTypePrivate, Meta: "unauthorized"}
	}

	err = s.ctrl.UpdateUserMessage(int64(idInt), req)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "User message updated successfully"}, nil
}

func (s *Server) deleteUserMessage(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	// Verify the message belongs to the user before deleting
	message, err := s.ctrl.GetUserMessage(int64(idInt))
	if err != nil {
		return nil, err
	}
	if message.ToUser != claim.UserId {
		return nil, gin.Error{Err: gin.Error{Err: nil}, Type: gin.ErrorTypePrivate, Meta: "unauthorized"}
	}

	err = s.ctrl.DeleteUserMessage(int64(idInt))
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "User message deleted successfully"}, nil
}

func (s *Server) selfHandleUserWs(ctx *gin.Context) {
	tok := ctx.Query("token")
	if tok == "" {
		httpx.WriteAuthErr(ctx, EmptyAuthTokenErr)
		return
	}
	claim, err := s.withAccessToken(tok)
	if err != nil {
		return
	}

	s.opt.CoreHub.HandleUserWS(claim.UserId, ctx)
}
