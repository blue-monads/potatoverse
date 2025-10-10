package engine

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
