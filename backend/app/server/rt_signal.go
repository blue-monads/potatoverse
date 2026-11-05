package server

import (
	"encoding/json"
	"strconv"

	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
)

// ListSignals lists all signals for a space/package
func (a *Server) ListSignals(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	spaceIdParam := ctx.Query("space_id")
	var spaceId int64
	if spaceIdParam != "" {
		spaceId, _ = strconv.ParseInt(spaceIdParam, 10, 64)
	}

	cond := make(map[any]any)
	if role := ctx.Query("role"); role != "" {
		cond["role"] = role
	}
	if signalKey := ctx.Query("signal_key"); signalKey != "" {
		cond["signal_key"] = signalKey
	}

	signals, err := a.ctrl.QuerySignals(installId, spaceId, cond)
	if err != nil {
		return nil, err
	}

	return signals, nil
}

// GetSignal gets a specific signal by ID
func (a *Server) GetSignal(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	signalId, err := strconv.ParseInt(ctx.Param("signalId"), 10, 64)
	if err != nil {
		return nil, err
	}

	signal, err := a.ctrl.GetSignalByID(installId, signalId)
	if err != nil {
		return nil, err
	}

	return signal, nil
}

// CreateSignal creates a new signal
func (a *Server) CreateSignal(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	var signalData *dbmodels.Signal
	if err := ctx.ShouldBindJSON(&signalData); err != nil {
		return nil, err
	}

	if signalData.EmitterInstallID == 0 {
		signalData.EmitterInstallID = installId
	}
	signalData.CreatedBy = claim.UserId

	signal, err := a.ctrl.CreateSignal(installId, signalData)
	if err != nil {
		return nil, err
	}

	return signal, nil
}

// UpdateSignal updates an existing signal
func (a *Server) UpdateSignal(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	signalId, err := strconv.ParseInt(ctx.Param("signalId"), 10, 64)
	if err != nil {
		return nil, err
	}

	var updateData map[string]any
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		return nil, err
	}

	signal, err := a.ctrl.UpdateSignalByID(installId, signalId, updateData)
	if err != nil {
		return nil, err
	}

	return signal, nil
}

// DeleteSignal deletes a signal
func (a *Server) DeleteSignal(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	signalId, err := strconv.ParseInt(ctx.Param("signalId"), 10, 64)
	if err != nil {
		return nil, err
	}

	err = a.ctrl.DeleteSignalByID(installId, signalId)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Signal deleted successfully"}, nil
}

// ListSignalEvents lists signal event execution history (targets)
func (a *Server) ListSignalEvents(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	signalId, _ := strconv.ParseInt(ctx.Query("signal_id"), 10, 64)
	status := ctx.Query("status")
	limit, _ := strconv.ParseInt(ctx.Query("limit"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.Query("offset"), 10, 64)

	if limit == 0 {
		limit = 100
	}

	targets, err := a.ctrl.QuerySignalTargets(installId, signalId, status, limit, offset)
	if err != nil {
		return nil, err
	}

	return targets, nil
}

type UpdateTargetStatusRequest struct {
	Status string `json:"status"` // blocked, new
	Reason string `json:"reason"`
}

// UpdateSignalTargetStatus updates the status of a signal target (e.g. block/unblock)
func (a *Server) UpdateSignalTargetStatus(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	targetId, err := strconv.ParseInt(ctx.Param("targetId"), 10, 64)
	if err != nil {
		return nil, err
	}

	var req UpdateTargetStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	err = a.ctrl.UpdateSignalTargetStatus(installId, targetId, req.Status, req.Reason)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Signal target status updated successfully"}, nil
}

type EmitSignalRequest struct {
	SignalKey      string         `json:"signal_key"`
	EmitterSpaceID int64          `json:"emitter_space_id"`
	Payload        any            `json:"payload"`
	Metadata       map[string]any `json:"metadata"`
}

// EmitSignal publishes a signal for testing/triggering
func (a *Server) EmitSignal(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	var req EmitSignalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	var payloadBytes []byte
	switch p := req.Payload.(type) {
	case string:
		payloadBytes = []byte(p)
	case []byte:
		payloadBytes = p
	default:
		payloadBytes, _ = json.Marshal(p)
	}

	opts := &xtypes.SignalOptions{
		SignalKey:        req.SignalKey,
		EmitterInstallId: installId,
		EmitterSpaceId:   req.EmitterSpaceID,
		Payload:          payloadBytes,
		Metadata:         req.Metadata,
	}

	err = a.ctrl.EmitSignal(installId, opts)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Signal emitted successfully"}, nil
}
