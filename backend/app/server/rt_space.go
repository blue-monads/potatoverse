package server

import (
	"fmt"
	"strconv"

	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/gin-gonic/gin"
)

// ListSpaceKV lists all KV entries for a space
func (a *Server) ListSpaceKV(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this space
	space, err := a.ctrl.GetSpace(installId)
	if err != nil {
		return nil, err
	}

	if space.OwnerID != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this space")
	}

	// Get query parameters for filtering
	groupName := ctx.Query("group")
	key := ctx.Query("key")
	tag1 := ctx.Query("tag1")
	tag2 := ctx.Query("tag2")
	tag3 := ctx.Query("tag3")

	// Build condition map
	cond := make(map[any]any)
	if groupName != "" {
		cond["group"] = groupName
	}
	if key != "" {
		cond["key"] = key
	}
	if tag1 != "" {
		cond["tag1"] = tag1
	}
	if tag2 != "" {
		cond["tag2"] = tag2
	}
	if tag3 != "" {
		cond["tag3"] = tag3
	}

	kvEntries, err := a.ctrl.QuerySpaceKV(installId, cond)
	if err != nil {
		return nil, err
	}

	return kvEntries, nil
}

// GetSpaceKV gets a specific KV entry by ID
func (a *Server) GetSpaceKV(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	kvId, err := strconv.ParseInt(ctx.Param("kvId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this space
	space, err := a.ctrl.GetSpace(installId)
	if err != nil {
		return nil, err
	}

	if space.OwnerID != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this space")
	}

	kvEntry, err := a.ctrl.GetSpaceKVByID(installId, kvId)
	if err != nil {
		return nil, err
	}

	return kvEntry, nil
}

// CreateSpaceKV creates a new KV entry
func (a *Server) CreateSpaceKV(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this space
	space, err := a.ctrl.GetSpace(installId)
	if err != nil {
		return nil, err
	}

	if space.OwnerID != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this space")
	}

	var kvData map[string]any
	if err := ctx.ShouldBindJSON(&kvData); err != nil {
		return nil, err
	}

	kvEntry, err := a.ctrl.CreateSpaceKV(installId, kvData)
	if err != nil {
		return nil, err
	}

	return kvEntry, nil
}

// UpdateSpaceKV updates an existing KV entry
func (a *Server) UpdateSpaceKV(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	kvId, err := strconv.ParseInt(ctx.Param("kvId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this space
	space, err := a.ctrl.GetSpace(installId)
	if err != nil {
		return nil, err
	}

	if space.OwnerID != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this space")
	}

	var updateData map[string]any
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		return nil, err
	}

	kvEntry, err := a.ctrl.UpdateSpaceKVByID(installId, kvId, updateData)
	if err != nil {
		return nil, err
	}

	return kvEntry, nil
}

// DeleteSpaceKV deletes a KV entry
func (a *Server) DeleteSpaceKV(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	kvId, err := strconv.ParseInt(ctx.Param("kvId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this space
	space, err := a.ctrl.GetSpace(installId)
	if err != nil {
		return nil, err
	}

	if space.OwnerID != claim.UserId {
		return nil, fmt.Errorf("you are not authorized to access this space")
	}

	err = a.ctrl.DeleteSpaceKVByID(installId, kvId)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "KV entry deleted successfully"}, nil
}
