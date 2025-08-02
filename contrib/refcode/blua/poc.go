package blua

type BLuaApp struct {
	Name        string         `json:"name"`
	Routes      []BLuaAppRoute `json:"routes"`
	Capabilites []string       `json:"capabilities"`
	Services    map[string]any `json:"services"`
}

type BLuaAppRoute struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"` // authed_http, http, ws
	Method   string            `json:"method"`
	Path     string            `json:"path"`
	Handlers map[string]string `json:"handlers"`
	Options  map[string]any    `json:"options"`
}
