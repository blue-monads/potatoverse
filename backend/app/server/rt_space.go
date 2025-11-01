package server

import (
	"strconv"

	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/gin-gonic/gin"
)

func (a *Server) ListSpaceKV(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

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

	// Parse pagination parameters
	offset := 0
	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	limit := 100 // default limit
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	kvEntries, err := a.ctrl.QuerySpaceKV(installId, cond, offset, limit)
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

	// fixme => permission check

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

	// fixme => permission check

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

	// fixme => permission check

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

	// fixme => permission check

	err = a.ctrl.DeleteSpaceKVByID(installId, kvId)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "KV entry deleted successfully"}, nil
}

// ListSpaceUsers lists all users for a space/package
func (a *Server) ListSpaceUsers(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

	// Get query parameters for filtering
	spaceIdParam := ctx.Query("space_id")
	userIdParam := ctx.Query("user_id")
	scopeParam := ctx.Query("scope")

	// Build condition map
	cond := make(map[any]any)
	if spaceIdParam != "" {
		spaceId, err := strconv.ParseInt(spaceIdParam, 10, 64)
		if err == nil {
			cond["space_id"] = spaceId
		}
	}
	if userIdParam != "" {
		userId, err := strconv.ParseInt(userIdParam, 10, 64)
		if err == nil {
			cond["user_id"] = userId
		}
	}
	if scopeParam != "" {
		cond["scope"] = scopeParam
	}

	users, err := a.ctrl.QuerySpaceUsers(installId, cond)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetSpaceUser gets a specific space user by ID
func (a *Server) GetSpaceUser(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	spaceUserId, err := strconv.ParseInt(ctx.Param("spaceUserId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

	spaceUser, err := a.ctrl.GetSpaceUserByID(installId, spaceUserId)
	if err != nil {
		return nil, err
	}

	return spaceUser, nil
}

// CreateSpaceUser creates a new space user
func (a *Server) CreateSpaceUser(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

	var userData map[string]any
	if err := ctx.ShouldBindJSON(&userData); err != nil {
		return nil, err
	}

	spaceUser, err := a.ctrl.CreateSpaceUser(installId, userData)
	if err != nil {
		return nil, err
	}

	return spaceUser, nil
}

// UpdateSpaceUser updates an existing space user
func (a *Server) UpdateSpaceUser(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	spaceUserId, err := strconv.ParseInt(ctx.Param("spaceUserId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

	var updateData map[string]any
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		return nil, err
	}

	spaceUser, err := a.ctrl.UpdateSpaceUserByID(installId, spaceUserId, updateData)
	if err != nil {
		return nil, err
	}

	return spaceUser, nil
}

// DeleteSpaceUser deletes a space user
func (a *Server) DeleteSpaceUser(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	spaceUserId, err := strconv.ParseInt(ctx.Param("spaceUserId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

	err = a.ctrl.DeleteSpaceUserByID(installId, spaceUserId)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Space user deleted successfully"}, nil
}

// ListEventSubscriptions lists all event subscriptions for a space/package
func (a *Server) ListEventSubscriptions(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

	// Get query parameters for filtering
	spaceIdParam := ctx.Query("space_id")
	eventKeyParam := ctx.Query("event_key")

	// Build condition map
	cond := make(map[any]any)
	if spaceIdParam != "" {
		spaceId, err := strconv.ParseInt(spaceIdParam, 10, 64)
		if err == nil {
			cond["space_id"] = spaceId
		}
	}
	if eventKeyParam != "" {
		cond["event_key"] = eventKeyParam
	}

	subscriptions, err := a.ctrl.QueryEventSubscriptions(installId, cond)
	if err != nil {
		return nil, err
	}

	return subscriptions, nil
}

// GetEventSubscription gets a specific event subscription by ID
func (a *Server) GetEventSubscription(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	subscriptionId, err := strconv.ParseInt(ctx.Param("subscriptionId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

	subscription, err := a.ctrl.GetEventSubscriptionByID(installId, subscriptionId)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// CreateEventSubscription creates a new event subscription
func (a *Server) CreateEventSubscription(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

	var subscriptionData map[string]any
	if err := ctx.ShouldBindJSON(&subscriptionData); err != nil {
		return nil, err
	}

	// Set created_by from claim if available
	if claim != nil && claim.UserId > 0 {
		subscriptionData["created_by"] = float64(claim.UserId)
	}

	subscription, err := a.ctrl.CreateEventSubscription(installId, subscriptionData)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// UpdateEventSubscription updates an existing event subscription
func (a *Server) UpdateEventSubscription(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	subscriptionId, err := strconv.ParseInt(ctx.Param("subscriptionId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

	var updateData map[string]any
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		return nil, err
	}

	subscription, err := a.ctrl.UpdateEventSubscriptionByID(installId, subscriptionId, updateData)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// DeleteEventSubscription deletes an event subscription
func (a *Server) DeleteEventSubscription(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	installId, err := strconv.ParseInt(ctx.Param("install_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	subscriptionId, err := strconv.ParseInt(ctx.Param("subscriptionId"), 10, 64)
	if err != nil {
		return nil, err
	}

	// fixme => permission check

	err = a.ctrl.DeleteEventSubscriptionByID(installId, subscriptionId)
	if err != nil {
		return nil, err
	}

	return gin.H{"message": "Event subscription deleted successfully"}, nil
}
