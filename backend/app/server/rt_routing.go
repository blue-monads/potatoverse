package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/blue-monads/potatoverse/backend/engine"
	"github.com/blue-monads/potatoverse/backend/services/datahub/dbmodels"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/gin-gonic/gin"
)

func (a *Server) routingRoutes(g *gin.RouterGroup) {
	g.GET("/root", a.withAccessTokenFn(a.adminGetRootRouting))
	g.POST("/root", a.withAccessTokenFn(a.adminUpdateRootRouting))
	g.POST("/root/reload", a.withAccessTokenFn(a.adminReloadRootRouting))
}

func (a *Server) checkAdminUser(claim *signer.AccessClaim) (*dbmodels.User, error) {
	if claim == nil || claim.UserId <= 0 {
		return nil, errors.New("unauthorized: missing access claim")
	}

	user, err := a.ctrl.GetUser(claim.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil || user.Ugroup != "admin" {
		return nil, errors.New("forbidden: requires admin usergroup")
	}

	return user, nil
}

func (a *Server) adminGetRootRouting(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	if _, err := a.checkAdminUser(claim); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return nil, nil
	}

	allSpaces, err := a.ctrl.Database().GetSpaceOps().ListSpaces()
	if err != nil {
		return nil, err
	}

	mainSpaces := make([]dbmodels.Space, 0)
	for _, s := range allSpaces {
		if s.SpaceType != "AppPlugin" && !strings.Contains(s.NamespaceKey, ":") {
			mainSpaces = append(mainSpaces, s)
		}
	}

	routes := make(map[string]*engine.RootRouteTarget)
	cfg, err := a.ctrl.Database().GetGlobalOps().GetGlobalConfig(engine.RootRoutingIndexConfigKey, engine.RootRoutingIndexConfigGroup)
	if err != nil {
		if !a.ctrl.Database().IsEmptyRowsError(err) {
			return nil, err
		}
	} else if cfg != nil && strings.TrimSpace(cfg.Value) != "" {
		_ = json.Unmarshal([]byte(cfg.Value), &routes)
	}

	if routes == nil {
		routes = make(map[string]*engine.RootRouteTarget)
	}

	return gin.H{
		"routes":      routes,
		"main_spaces": mainSpaces,
	}, nil
}

type UpdateRootRoutingReq struct {
	Routes map[string]*engine.RootRouteTarget `json:"routes"`
}

func (a *Server) adminUpdateRootRouting(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	if _, err := a.checkAdminUser(claim); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return nil, nil
	}

	var req UpdateRootRoutingReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	cleanedRoutes := make(map[string]*engine.RootRouteTarget)

	for pattern, target := range req.Routes {
		p := strings.ToLower(strings.TrimSpace(pattern))
		if p == "" {
			continue
		}
		if target == nil || target.SpaceId <= 0 {
			return nil, fmt.Errorf("invalid space_id for pattern '%s'", pattern)
		}

		// Ensure space exists and is a main space (not an AppPlugin and not a subkey space)
		space, err := a.ctrl.Database().GetSpaceOps().GetSpace(target.SpaceId)
		if err != nil || space == nil {
			return nil, fmt.Errorf("space ID %d not found for pattern '%s'", target.SpaceId, pattern)
		}

		if space.SpaceType == "AppPlugin" || strings.Contains(space.NamespaceKey, ":") {
			return nil, fmt.Errorf("space '%s' (ID %d) is not a main space; only main spaces are allowed in root routing", space.NamespaceKey, space.ID)
		}

		cleanedRoutes[p] = &engine.RootRouteTarget{
			SpaceId: target.SpaceId,
		}
	}

	bytes, err := json.Marshal(cleanedRoutes)
	if err != nil {
		return nil, fmt.Errorf("failed to encode routing index: %w", err)
	}

	jsonStr := string(bytes)

	_, err = a.ctrl.Database().GetGlobalOps().GetGlobalConfig(engine.RootRoutingIndexConfigKey, engine.RootRoutingIndexConfigGroup)
	if err != nil && a.ctrl.Database().IsEmptyRowsError(err) {
		_, err = a.ctrl.Database().GetGlobalOps().AddGlobalConfig(&dbmodels.GlobalConfig{
			Key:       engine.RootRoutingIndexConfigKey,
			GroupName: engine.RootRoutingIndexConfigGroup,
			Value:     jsonStr,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert routing index: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to check existing routing config: %w", err)
	} else {
		err = a.ctrl.Database().GetGlobalOps().UpdateGlobalConfigByKey(
			engine.RootRoutingIndexConfigKey,
			engine.RootRoutingIndexConfigGroup,
			map[string]any{"value": jsonStr},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update routing index: %w", err)
		}
	}

	if a.engine != nil {
		if err := a.engine.ReloadRootRoutingIndex(); err != nil {
			return nil, fmt.Errorf("saved, but reload failed: %w", err)
		}
	}

	return gin.H{
		"message": "Root routing index updated and reloaded successfully",
		"routes":  cleanedRoutes,
	}, nil
}

func (a *Server) adminReloadRootRouting(claim *signer.AccessClaim, ctx *gin.Context) (any, error) {
	if _, err := a.checkAdminUser(claim); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return nil, nil
	}

	if a.engine != nil {
		if err := a.engine.ReloadRootRoutingIndex(); err != nil {
			return nil, fmt.Errorf("failed to reload root routing index: %w", err)
		}
	}

	return gin.H{
		"message": "Root routing index reloaded successfully",
	}, nil
}
