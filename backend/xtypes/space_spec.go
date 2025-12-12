package xtypes

type PotatoSpec struct {
	Specs  map[string]*SpaceSpec `json:"specs"`
	Models []ModelSpec           `json:"models"`
}

type SpaceSpec struct {
	Scopes []ScopeSpec   `json:"permissions"`
	Events []EventSpec   `json:"events"`
	Slots  []HandlerSpec `json:"slots"`
	APIs   []HandlerSpec `json:"apis"`
}

type ModelSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Fields      map[string]any `json:"schema"`
}

type ScopeSpec struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type EventSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Schema      map[string]any `json:"schema"`
	SchemaFile  string         `json:"schema_file"`
}

type HandlerSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Schema      map[string]any `json:"schema"`
	SchemaFile  string         `json:"schema_file"`
}
