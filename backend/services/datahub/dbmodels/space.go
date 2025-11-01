package dbmodels

import "time"

type Space struct {
	ID                int64  `json:"id" db:"id,omitempty"`
	InstalledId       int64  `json:"install_id" db:"install_id,omitempty"`
	NamespaceKey      string `json:"namespace_key" db:"namespace_key,omitempty"`
	ExecutorType      string `json:"executor_type" db:"executor_type,omitempty"`
	SubType           string `json:"sub_type" db:"sub_type,omitempty"`
	RouteOptions      string `json:"route_options" db:"route_options,omitempty"`
	McpEnabled        bool   `json:"mcp_enabled" db:"mcp_enabled,omitempty"`
	McpDefinitionFile string `json:"mcp_definition_file" db:"mcp_definition_file,omitempty"`
	McpOptions        string `json:"mcp_options" db:"mcp_options,omitempty"`
	DevServePort      int64  `json:"dev_serve_port" db:"dev_serve_port,omitempty"`
	ServerFile        string `json:"server_file" db:"server_file,omitempty"`

	DevMode bool `json:"dev_mode" db:"dev_mode,omitempty"`

	OverlayForSpaceID int64  `json:"overlay_for_space_id" db:"overlay_for_space_id,omitempty"`
	OwnerID           int64  `json:"owned_by" db:"owned_by"`
	ExtraMeta         string `json:"extrameta" db:"extrameta,omitempty"`
	IsInitilized      bool   `json:"is_initilized" db:"is_initilized,omitempty"`
	IsPublic          bool   `json:"is_public" db:"is_public,omitempty"`
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
	Group   string `json:"group" db:"group"`
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

type SpaceKV struct {
	ID      int64  `json:"id" db:"id,omitempty"`
	Key     string `json:"key" db:"key"`
	Group   string `json:"group" db:"group"`
	Value   string `json:"value" db:"value"`
	SpaceID int64  `json:"space_id" db:"space_id"`
	Tag1    string `json:"tag1" db:"tag1,omitempty"`
	Tag2    string `json:"tag2" db:"tag2,omitempty"`
	Tag3    string `json:"tag3" db:"tag3,omitempty"`
}
