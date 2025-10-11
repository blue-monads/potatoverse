package engine

import (
	"encoding/json"
	"fmt"

	"github.com/blue-monads/turnix/backend/xtypes/models"
)

type SpaceRouteIndexItem struct {
	packageId         int64
	spaceId           int64
	overlayForSpaceId int64
	routeOption       models.PotatoRouteOptions
}

type PluginRouteIndexItem struct {
	pluginId    int64
	packageId   int64
	spaceId     int64
	routeOption models.PotatoRouteOptions
}

func (e *Engine) LoadRoutingIndex() error {

	nextRoutingIndex := make(map[string]*SpaceRouteIndexItem)

	spaces, err := e.db.ListSpaces()
	if err != nil {
		return err
	}

	for _, space := range spaces {

		routeOptions := models.PotatoRouteOptions{}
		err = json.Unmarshal([]byte(space.RouteOptions), &routeOptions)
		if err != nil {
			routeOptions.ServeFolder = "public"
			routeOptions.TrimPathPrefix = ""
			routeOptions.ForceHtmlExtension = false
			routeOptions.ForceIndexHtmlFile = true

			e.logger.Warn("error unmarshalling route options, using default values", "error", err, "space_id", space.ID, "route_options", space.RouteOptions)

		}

		indexItem := &SpaceRouteIndexItem{
			packageId:   space.PackageID,
			spaceId:     space.ID,
			routeOption: routeOptions,
		}

		if space.OwnsNamespace {
			nextRoutingIndex[space.NamespaceKey] = indexItem
		} else {
			nextRoutingIndex[fmt.Sprintf("%d|_|%s", space.ID, space.NamespaceKey)] = indexItem
		}
	}

	e.riLock.Lock()
	e.RoutingIndex = nextRoutingIndex
	e.riLock.Unlock()

	return nil
}

func (e *Engine) getIndex(spaceKey string, spaceId int64) *SpaceRouteIndexItem {
	e.riLock.RLock()
	defer e.riLock.RUnlock()

	if spaceId != 0 {
		spaceKey = fmt.Sprintf("%d|_|%s", spaceId, spaceKey)
	}

	return e.RoutingIndex[spaceKey]
}
