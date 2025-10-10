package engine

type SpaceRouteIndexItem struct {
	packageId         int64
	spaceId           int64
	overlayForSpaceId int64
	serveFolder       string
	trimPathPrefix    string
}

type PluginRouteIndexItem struct {
	pluginId       int64
	packageId      int64
	spaceId        int64
	trimPathPrefix string
	serveFolder    string
}
