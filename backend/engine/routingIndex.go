package engine

import "fmt"

type SpaceRouteIndexItem struct {
	packageId         int64
	spaceId           int64
	overlayForSpaceId int64
	routeOption       RouteOption
}

type PluginRouteIndexItem struct {
	pluginId    int64
	packageId   int64
	spaceId     int64
	routeOption RouteOption
}

type RouteOption struct {
	ServeFolder        string
	TrimPathPrefix     string
	ForceHtmlExtension bool
	ForceIndexHtmlFile bool
}

func (e *Engine) LoadRoutingIndex() error {

	nextRoutingIndex := make(map[string]*SpaceRouteIndexItem)

	spaces, err := e.db.ListSpaces()
	if err != nil {
		return err
	}

	for _, space := range spaces {
		indexItem := &SpaceRouteIndexItem{
			packageId: space.PackageID,
			spaceId:   space.ID,
			routeOption: RouteOption{
				ServeFolder:        "public",
				TrimPathPrefix:     "",
				ForceHtmlExtension: false,
				ForceIndexHtmlFile: false,
			},
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
