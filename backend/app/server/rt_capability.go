package server

import (
	"strconv"

	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/gin-gonic/gin"
)

// ListSpaceCapabilities lists all capabilities for a package/space
// Supports both package-level (space_id=0) and space-level (space_id>0) capabilities
func (a *Server) ListSpaceCapabilities(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

	// Get query parameters for filtering
	capabilityType := ctx.Query("capability_type")
	spaceIdStr := ctx.Query("space_id")
	
	// If space_id not provided, show all (both package-level and space-level)
	// If space_id=0 is explicitly provided, show only package-level
	// If space_id>0 is provided, show only that space's capabilities

	// Build condition map
	cond := make(map[any]any)
	if capabilityType != "" {
		cond["capability_type"] = capabilityType
	}
	if spaceIdStr != "" {
		sid, err := strconv.ParseInt(spaceIdStr, 10, 64)
		if err == nil {
			cond["space_id"] = sid
		}
	}

	capabilities, err := a.ctrl.QuerySpaceCapabilities(installId, cond)
	if err != nil {
		return nil, err
	}

	return capabilities, nil
}

// GetSpaceCapability gets a specific capability by ID
func (a *Server) GetSpaceCapability(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	capabilityId, err := strconv.ParseInt(ctx.Param("capabilityId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

	capability, err := a.ctrl.GetSpaceCapabilityByID(installId, capabilityId)
	if err != nil {
		return nil, err
	}

	return capability, nil
}

// CreateSpaceCapability creates a new capability
// Can be created at package level (space_id=0 or omitted) or space level (space_id>0)
func (a *Server) CreateSpaceCapability(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

	var capabilityData map[string]any
	if err := ctx.ShouldBindJSON(&capabilityData); err != nil {
		return nil, err
	}

	// If space_id is not provided, default to 0 (package-level)
	if _, ok := capabilityData["space_id"]; !ok {
		capabilityData["space_id"] = float64(0)
	}

	capability, err := a.ctrl.CreateSpaceCapability(installId, capabilityData)
	if err != nil {
		return nil, err
	}

	return capability, nil
}

// UpdateSpaceCapability updates an existing capability
func (a *Server) UpdateSpaceCapability(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	capabilityId, err := strconv.ParseInt(ctx.Param("capabilityId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

	var updateData map[string]any
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		return nil, err
	}

	capability, err := a.ctrl.UpdateSpaceCapabilityByID(installId, capabilityId, updateData)
	if err != nil {
		return nil, err
	}

	return capability, nil
}

// DeleteSpaceCapability deletes a capability
func (a *Server) DeleteSpaceCapability(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	capabilityId, err := strconv.ParseInt(ctx.Param("capabilityId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

	err = a.ctrl.DeleteSpaceCapabilityByID(installId, capabilityId)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Capability deleted successfully"}, nil
}

// ListCapabilityTypes lists all available capability type definitions
func (a *Server) ListCapabilityTypes(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	definitions := a.engine.GetCapabilityDefinitions()
	return definitions, nil
}
