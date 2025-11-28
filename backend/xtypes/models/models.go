package models

type PotatoPackage struct {
	Name          string           `json:"name" toml:"name"`
	Slug          string           `json:"slug" toml:"slug"`
	Info          string           `json:"info" toml:"info"`
	CanonicalUrl  string           `json:"canonical_url" toml:"canonical_url"`
	Tags          []string         `json:"tags" toml:"tags"`
	FormatVersion string           `json:"format_version" toml:"format_version"`
	AuthorName    string           `json:"author_name" toml:"author_name"`
	AuthorEmail   string           `json:"author_email" toml:"author_email"`
	AuthorSite    string           `json:"author_site" toml:"author_site"`
	SourceCode    string           `json:"source_code" toml:"source_code"`
	License       string           `json:"license" toml:"license"`
	Version       string           `json:"version" toml:"version"`
	Artifacts     []PotatoArtifact `json:"artifacts" toml:"artifacts"`
	InitPage      string           `json:"init_page" toml:"init_page"`
	UpdatePage    string           `json:"update_page" toml:"update_page"`

	// for local dev
	Developer *DeveloperOptions `json:"developer,omitempty" toml:"developer,omitempty"`
}

type DeveloperOptions struct {
	ServerUrl     string   `json:"server_url" toml:"server_url"`
	Token         string   `json:"token" toml:"token"`
	OutputZipFile string   `json:"output_zip_file,omitempty" toml:"output_zip_file,omitempty"`
	IncludeFiles  []string `json:"include_files,omitempty" toml:"include_files,omitempty"`
	ExcludeFiles  []string `json:"exclude_files,omitempty" toml:"exclude_files,omitempty"`
}

type PotatoArtifact struct {
	Namespace       string             `json:"namespace" toml:"namespace"`
	Kind            string             `json:"kind" toml:"kind"`
	ExecutorType    string             `json:"executor_type" toml:"executor_type"`
	ExecutorSubType string             `json:"executor_sub_type" toml:"executor_sub_type"`
	ServerFile      string             `json:"server_file" toml:"server_file"`
	RouteOptions    PotatoRouteOptions `json:"route_options" toml:"route_options"`
	McpOptions      PotatoMcpOptions   `json:"mcp_options" toml:"mcp_options"`
	DevServePort    int                `json:"dev_serve_port" toml:"dev_serve_port"`
}

type PotatoRouteOptions struct {
	RouterType         string        `json:"router_type" toml:"router_type"`
	ServeFolder        string        `json:"serve_folder" toml:"serve_folder"`
	ForceHtmlExtension bool          `json:"force_html_extension" toml:"force_html_extension"`
	ForceIndexHtmlFile bool          `json:"force_index_html_file" toml:"force_index_html_file"`
	TrimPathPrefix     string        `json:"trim_path_prefix" toml:"trim_path_prefix"`
	TemplateFolder     string        `json:"template_folder" toml:"template_folder"`
	Routes             []PotatoRoute `json:"routes" toml:"routes"`
}

type PotatoRoute struct {
	Path    string `json:"path" toml:"path"`
	Method  string `json:"method" toml:"method"`
	Type    string `json:"type" toml:"type"`
	Handler string `json:"handler" toml:"handler"`
	File    string `json:"file" toml:"file"`
}

type PotatoMcpOptions struct {
	Enabled        bool   `json:"enabled" toml:"enabled"`
	DefinitionFile string `json:"definition_file" toml:"definition_file"`
}

type PotatoDevOptions struct {
	ServerUrl string `json:"server_url" toml:"server_url"`
	Token     string `json:"token" toml:"token"`
}
