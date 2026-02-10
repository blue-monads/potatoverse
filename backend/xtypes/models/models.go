package models

type PotatoPackage struct {
	Name         string             `json:"name" yaml:"name"`
	Slug         string             `json:"slug" yaml:"slug"`
	Info         string             `json:"info" yaml:"info"`
	SpecialPages map[string]string  `json:"special_pages" yaml:"special_pages,omitempty"`
	Spaces       []PotatoSpace      `json:"spaces" yaml:"spaces,omitempty"`
	Capabilities []PotatoCapability `json:"capabilities" yaml:"capabilities,omitempty"`
	CanonicalUrl string             `json:"canonical_url" yaml:"canonical_url"`

	FormatVersion string `json:"format_version" yaml:"format_version"`
	AuthorName    string `json:"author_name" yaml:"author_name"`
	AuthorEmail   string `json:"author_email" yaml:"author_email"`
	AuthorSite    string `json:"author_site" yaml:"author_site"`
	SourceCode    string `json:"source_code" yaml:"source_code"`
	License       string `json:"license" yaml:"license"`
	Version       string `json:"version" yaml:"version"`

	Tags []string `json:"tags" yaml:"tags"`

	// for local dev
	Developer *DeveloperOptions `json:"developer,omitempty" yaml:"developer,omitempty"`
}

type DeveloperOptions struct {
	ServerUrl     string   `json:"server_url" yaml:"server_url"`
	Token         string   `json:"token" yaml:"token"`
	TokenEnv      string   `json:"token_env" yaml:"token_env"`
	OutputZipFile string   `json:"output_zip_file,omitempty" yaml:"output_zip_file,omitempty"`
	IncludeFiles  []string `json:"include_files,omitempty" yaml:"include_files,omitempty"`
	ExcludeFiles  []string `json:"exclude_files,omitempty" yaml:"exclude_files,omitempty"`
	BuildCommand  string   `json:"build_command" yaml:"build_command"`
}

type PotatoCapability struct {
	Name    string         `json:"name" yaml:"name"`
	Type    string         `json:"type" yaml:"type"`
	Options map[string]any `json:"options" yaml:"options"`
	Spaces  []string       `json:"spaces" yaml:"spaces"`
}

type PotatoSpace struct {
	Namespace       string             `json:"namespace" yaml:"namespace"`
	ExecutorType    string             `json:"executor_type" yaml:"executor_type"`
	ExecutorSubType string             `json:"executor_sub_type" yaml:"executor_sub_type"`
	ServerFile      string             `json:"server_file" yaml:"server_file"`
	RouteOptions    PotatoRouteOptions `json:"route_options" yaml:"route_options"`
	DevServePort    int                `json:"dev_serve_port" yaml:"dev_serve_port"`
	IsDefault       bool               `json:"is_default" yaml:"is_default"`
}

type PotatoRouteOptions struct {
	RouterType         string        `json:"router_type" yaml:"router_type"`
	ServeFolder        string        `json:"serve_folder" yaml:"serve_folder"`
	ForceHtmlExtension bool          `json:"force_html_extension" yaml:"force_html_extension"`
	ForceIndexHtmlFile bool          `json:"force_index_html_file" yaml:"force_index_html_file"`
	OnNotFoundFile     string        `json:"on_not_found_file" yaml:"on_not_found_file"`
	TrimPathPrefix     string        `json:"trim_path_prefix" yaml:"trim_path_prefix"`
	TemplateFolder     string        `json:"template_folder" yaml:"template_folder"`
	Routes             []PotatoRoute `json:"routes" yaml:"routes"`
}

type PotatoRoute struct {
	Path    string `json:"path" yaml:"path"`
	Method  string `json:"method" yaml:"method"`
	Type    string `json:"type" yaml:"type"`
	Handler string `json:"handler" yaml:"handler"`
	File    string `json:"file" yaml:"file"`
}

type PotatoDevOptions struct {
	ServerUrl string `json:"server_url" yaml:"server_url"`
	Token     string `json:"token" yaml:"token"`
}
