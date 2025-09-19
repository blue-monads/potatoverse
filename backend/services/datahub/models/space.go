package models

import "time"

type Space struct {
	ID            int64  `json:"id" db:"id,omitempty"`
	NamespaceKey  string `json:"namespace_key" db:"namespace_key,omitempty"`
	OwnsNamespace bool   `json:"owns_namespace" db:"owns_namespace,omitempty"`
	PackageID     int64  `json:"package_id" db:"package_id,omitempty"`
	ExecutorType  string `json:"executor_type" db:"executor_type,omitempty"`
	SubType       string `json:"sub_type" db:"sub_type,omitempty"`
	OwnerID       int64  `json:"owned_by" db:"owned_by"`
	ExtraMeta     string `json:"extrameta" db:"extrameta,omitempty"`
	IsInitilized  bool   `json:"is_initilized" db:"is_initilized,omitempty"`
	IsPublic      bool   `json:"is_public" db:"is_public,omitempty"`
}

type SpaceUser struct {
	ID        int64  `json:"id" db:"id,omitempty"`
	UserID    int64  `json:"user_id" db:"userId"`
	SpaceID   int64  `json:"space_id" db:"spaceId"`
	Scope     string `json:"scope" db:"scope,omitempty"`
	Token     string `json:"token" db:"token"`
	ExtraMeta string `json:"extrameta" db:"extrameta,omitempty"`
}

type SpaceTypes struct {
	Name        string   `json:"name"`
	Ptype       string   `json:"ptype"`
	Slug        string   `json:"slug"`
	Info        string   `json:"info"`
	Icon        string   `json:"icon"`
	IsExternal  bool     `json:"is_external"`
	BaseLink    string   `json:"base_link,omitempty"`
	LinkPattern string   `json:"link_pattern,omitempty"`
	EventTypes  []string `json:"event_types,omitempty"`
}

type PluginImport struct {
	Name            string `json:"name" yaml:"name"`
	AppType         string `json:"apptype" yaml:"apptype"`
	ProjectTypeSlug string `json:"project_type_slug" yaml:"project_type_slug"`
	ServerCode      string `json:"server_code" yaml:"server_code"`
	ClientCode      string `json:"client_code" yaml:"client_code"`
}

type SpacePlugin struct {
	ID         int64      `json:"id" db:"id,omitempty"`
	Name       string     `json:"name" db:"name"`
	Type       string     `json:"ptype" db:"ptype"`
	SpaceID    int64      `json:"space_id" db:"space_id"`
	ServerCode string     `json:"server_code" db:"server_code"`
	ClientCode string     `json:"client_code" db:"client_code"`
	CreatedBy  int64      `json:"created_by" db:"created_by"`
	UpdatedBy  int64      `json:"updated_by" db:"updated_by"`
	CreatedAt  *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at" db:"updated_at"`
}

type SpaceConfig struct {
	ID      int64  `json:"id" db:"id,omitempty"`
	Key     string `json:"key" db:"key"`
	Group   string `json:"group_name" db:"group_name"`
	Value   string `json:"value" db:"value"`
	SpaceID int64  `json:"space_id" db:"space_id"`
}

type SpaceTableColumn struct {
	Cid          int64  `json:"id" db:"id,omitempty"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	NotNull      bool   `json:"not_null"`
	DefaultValue string `json:"default_value"`
	PrimaryKey   bool   `json:"primary_key"`
}

// Manifest

type SpaceManifest struct {
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Info        string         `json:"info"`
	Type        string         `json:"type"`
	Format      string         `json:"format"`
	Tags        []string       `json:"tags"`
	Routes      []Route        `json:"routes"`
	LinkPattern string         `json:"link_pattern"`
	ServerFile  string         `json:"server_file"`
	Services    map[string]any `json:"services"`
	ServeFolder string         `json:"serve_folder"`
}

type Route struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"` // authed_http, http, ws
	Method  string         `json:"method"`
	Path    string         `json:"path"`
	Handler string         `json:"handler"`
	Options map[string]any `json:"options"`
}

type SpaceKV struct {
	ID      int64  `json:"id" db:"id,omitempty" luautil:"-"`
	Key     string `json:"key" db:"key" luautil:"key"`
	Group   string `json:"group_name" db:"group_name" luautil:"group"`
	Value   string `json:"value" db:"value" luautil:"value"`
	SpaceID int64  `json:"space_id" db:"space_id"`
	Tag1    string `json:"tag1" db:"tag1,omitempty" luautil:"tag1"`
	Tag2    string `json:"tag2" db:"tag2,omitempty" luautil:"tag2"`
	Tag3    string `json:"tag3" db:"tag3,omitempty" luautil:"tag3"`
}
