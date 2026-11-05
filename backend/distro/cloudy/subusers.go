package cloudy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/blue-monads/potatoverse/backend/app/actions"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	xutils "github.com/blue-monads/potatoverse/backend/utils"
	"github.com/gin-gonic/gin"
)

type addSubUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	UGroup   string `json:"ugroup"`
}

type setSubUserDisabledRequest struct {
	Disabled bool `json:"disabled"`
}

func (a *CloudyApp) handleListSubUsers(c *gin.Context) {
	ctrl, tenant, err := a.subAppController(c)
	if err != nil {
		writeSubUserErr(c, err)
		return
	}

	users, err := ctrl.ListUsers(0, 500)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]gin.H, 0, len(users))
	for i := range users {
		out = append(out, subUserJSON(&users[i]))
	}
	c.JSON(http.StatusOK, gin.H{
		"tenant": tenant,
		"users":  out,
	})
}

func (a *CloudyApp) handleAddSubUser(c *gin.Context) {
	ctrl, tenant, err := a.subAppController(c)
	if err != nil {
		writeSubUserErr(c, err)
		return
	}

	var req addSubUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ugroup := strings.ToLower(strings.TrimSpace(req.UGroup))
	if ugroup == "" {
		ugroup = "normal"
	}
	if ugroup != "admin" && ugroup != "normal" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ugroup must be admin or normal"})
		return
	}

	user, err := ctrl.AddUserDirect(req.Name, req.Password, req.Email, ugroup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"tenant": tenant,
		"user":   subUserJSON(user),
	})
}

func (a *CloudyApp) handleResetSubUserPassword(c *gin.Context) {
	ctrl, tenant, err := a.subAppController(c)
	if err != nil {
		writeSubUserErr(c, err)
		return
	}

	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ctrl.GetUser(id)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	hash, err := xutils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.SetUserPassword(id, hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "tenant": tenant, "user_id": id})
}

func (a *CloudyApp) handleSetSubUserDisabled(c *gin.Context) {
	ctrl, tenant, err := a.subAppController(c)
	if err != nil {
		writeSubUserErr(c, err)
		return
	}

	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req setSubUserDisabledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := ctrl.GetUser(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if req.Disabled {
		err = ctrl.DeactivateUser(id)
	} else {
		err = ctrl.ActivateUser(id)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":       true,
		"tenant":   tenant,
		"user_id":  id,
		"disabled": req.Disabled,
	})
}

func (a *CloudyApp) subAppController(c *gin.Context) (*actions.Controller, string, error) {
	tenant, err := a.resolveSubUserTenant(c)
	if err != nil {
		return nil, "", err
	}

	sub, err := a.ensureSubApp(tenant)
	if err != nil {
		return nil, tenant, err
	}

	ctrl, err := sub.Controller()
	if err != nil {
		return nil, tenant, err
	}
	return ctrl, tenant, nil
}

func (a *CloudyApp) resolveSubUserTenant(c *gin.Context) (string, error) {
	claim := getClaim(c)
	if claim == nil {
		return "", errSubUserUnauthorized
	}

	raw := strings.TrimSpace(c.Query("for_tenant_id"))
	if raw == "" {
		raw = strings.TrimSpace(c.Query("for_slug"))
	}

	if raw == "" {
		if claim.Slug != "" {
			return claim.Slug, nil
		}
		inst, _, err := a.getPrimaryInstanceForUser(claim.UserID)
		if err != nil || inst == nil {
			return "", fmt.Errorf("no instance on this account")
		}
		return inst.Slug, nil
	}

	if claim.UType != UTypeAdmin {
		return "", errSubUserAdminRequired
	}

	// Admin can pass either numeric ID or slug string
	if a.slugExists(raw) {
		return raw, nil
	}

	id, err := parseIDParam(raw)
	if err != nil || id <= 0 {
		return "", fmt.Errorf("invalid for_tenant_id or slug")
	}

	inst, _, err := a.getPrimaryInstanceForUser(id)
	if err != nil || inst == nil {
		return "", fmt.Errorf("instance not found for user")
	}
	return inst.Slug, nil
}

func subUserJSON(u *dbmodels.User) gin.H {
	if u == nil {
		return gin.H{}
	}
	username := ""
	if u.Username != nil {
		username = *u.Username
	}
	return gin.H{
		"id":          u.ID,
		"name":        u.Name,
		"email":       u.Email,
		"username":    username,
		"ugroup":      u.Ugroup,
		"utype":       u.Utype,
		"is_verified": u.IsVerified,
		"disabled":    u.Disabled,
	}
}

func writeSubUserErr(c *gin.Context, err error) {
	switch err {
	case errSubUserUnauthorized:
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errSubUserAdminRequired:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		msg := err.Error()
		status := http.StatusBadRequest
		if strings.Contains(msg, "disabled") {
			status = http.StatusForbidden
		} else if strings.Contains(msg, "not found") || strings.Contains(msg, "not registered") {
			status = http.StatusNotFound
		} else if strings.Contains(msg, "not loaded") || strings.Contains(msg, "not verified") {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": msg})
	}
}

var (
	errSubUserUnauthorized  = fmt.Errorf("unauthorized")
	errSubUserAdminRequired = fmt.Errorf("admin required to specify another tenant or instance")
)
