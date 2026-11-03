package rroute

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"strings"
	"sync"

	"github.com/blue-monads/potatoverse/backend/engine/erouter/rspace"
	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/utils/libx/httpx"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/gin-gonic/gin"
)

/*

RootRoutingIndex { <domain: string> => <space_id: int64> }
* => { space_id: 12 }
example.com => { space_id: 11 }
ee.example => { space_id: 14 }
localhost => { space_id: 15 }
*.doomsday.com => { space_id: 13 }


GlobalConfig
(ROOT_ROUTING_INDEX, SYSTEM) = <json>


then call space

*/

const (
	RootRoutingIndexConfigKey   = "ROOT_ROUTING_INDEX"
	RootRoutingIndexConfigGroup = "SYSTEM"
)

type RootRouteTarget struct {
	SpaceId int64 `json:"space_id"`
}

type RootRouter struct {
	engine MyEngine
	logger *slog.Logger

	indexLock sync.RWMutex
	routes    map[string]*RootRouteTarget
}

type MyEngine interface {
	GetDatabase() datahub.Database
	GetSpaceRouter() *rspace.SpaceRouter
}

func NewRootRouter(e MyEngine, logger *slog.Logger) *RootRouter {
	return &RootRouter{
		engine: e,
		routes: make(map[string]*RootRouteTarget),
	}
}

func (r *RootRouter) Serve(ctx *gin.Context) {
	spaceId, found := r.MatchDomain(ctx.Request.Host)
	if !found || spaceId == 0 {
		qq.Println("@RootRouter/not_found", ctx.Request.Host)
		httpx.NotFound(ctx)
		return
	}

	qq.Println("@RootRouter/matched", ctx.Request.Host, spaceId)
	if sr := r.engine.GetSpaceRouter(); sr != nil {
		sr.ServeSpaceFile(ctx, spaceId)
	}
}

func (r *RootRouter) MatchDomain(rawHost string) (int64, bool) {
	host := strings.ToLower(strings.TrimSpace(rawHost))
	if host == "" {
		return 0, false
	}

	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	} else if idx := strings.Index(host, ":"); idx != -1 {
		hostname = host[:idx]
	}
	hostname = strings.TrimSpace(hostname)

	// Grab map reference under read lock, then unlock immediately (map is immutable after reload)
	r.indexLock.RLock()
	routes := r.routes
	r.indexLock.RUnlock()

	if len(routes) == 0 {
		return 0, false
	}

	// 1. Exact match with host (including port if configured, e.g. "localhost:8080")
	if target, ok := routes[host]; ok && target != nil && target.SpaceId > 0 {
		return target.SpaceId, true
	}

	// 2. Exact match with hostname (e.g. "example.com", "localhost")
	if target, ok := routes[hostname]; ok && target != nil && target.SpaceId > 0 {
		return target.SpaceId, true
	}

	// 3. Trim domain prefix and try with * (e.g. "sub.doomsday.com" -> "*.doomsday.com")
	curr := hostname
	for {
		dot := strings.Index(curr, ".")
		if dot == -1 {
			break
		}
		wildcard := "*" + curr[dot:]
		if target, ok := routes[wildcard]; ok && target != nil && target.SpaceId > 0 {
			return target.SpaceId, true
		}
		curr = curr[dot+1:]
	}

	// Also check apex fallback (e.g. "doomsday.com" -> "*.doomsday.com")
	if target, ok := routes["*."+hostname]; ok && target != nil && target.SpaceId > 0 {
		return target.SpaceId, true
	}

	// 4. Catch-all fallback "*"
	if target, ok := routes["*"]; ok && target != nil && target.SpaceId > 0 {
		return target.SpaceId, true
	}

	return 0, false
}

func (r *RootRouter) ReloadIndex() error {
	if r.engine == nil {
		return errors.New("engine or database not initialized")
	}

	db := r.engine.GetDatabase()
	if db == nil {
		return errors.New("database not available")
	}

	globalOps := db.GetGlobalOps()
	if globalOps == nil {
		return errors.New("global operations not available")
	}

	config, err := globalOps.GetGlobalConfig(RootRoutingIndexConfigKey, RootRoutingIndexConfigGroup)
	if err != nil {
		if db.IsEmptyRowsError(err) {
			r.setRoutes(nil)
			return nil
		}
		return err
	}

	if config == nil || strings.TrimSpace(config.Value) == "" {
		r.setRoutes(nil)
		return nil
	}

	var rawRoutes map[string]*RootRouteTarget
	err = json.Unmarshal([]byte(config.Value), &rawRoutes)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to unmarshal ROOT_ROUTING_INDEX", "error", err)
		}
		return err
	}

	r.setRoutes(rawRoutes)
	return nil
}

func (r *RootRouter) setRoutes(rawRoutes map[string]*RootRouteTarget) {
	routes := make(map[string]*RootRouteTarget, len(rawRoutes))
	for pattern, target := range rawRoutes {
		if target == nil || target.SpaceId <= 0 {
			continue
		}

		p := strings.ToLower(strings.TrimSpace(pattern))
		routes[p] = target
	}

	r.indexLock.Lock()
	r.routes = routes
	r.indexLock.Unlock()
}

func (r *RootRouter) GetRoutes() map[string]int64 {
	r.indexLock.RLock()
	routes := r.routes
	r.indexLock.RUnlock()

	result := make(map[string]int64, len(routes))
	for k, v := range routes {
		if v != nil {
			result[k] = v.SpaceId
		}
	}
	return result
}
