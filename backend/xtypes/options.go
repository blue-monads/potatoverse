package xtypes

type AppOptions struct {
	Name         string            `json:"name,omitempty" toml:"name,omitempty"`
	Port         int               `json:"port,omitempty" toml:"port,omitempty"`
	Hosts        []Host            `json:"hosts,omitempty" toml:"hosts,omitempty"`
	MasterSecret string            `json:"master_secret,omitempty" toml:"master_secret,omitempty"`
	Debug        bool              `json:"debug_mode,omitempty" toml:"debug_mode,omitempty"`
	WorkingDir   string            `json:"working_dir,omitempty" toml:"working_dir,omitempty"`
	SocketFile   string            `json:"socket_file,omitempty" toml:"socket_file,omitempty"`
	Mailer       MailerOptions     `json:"mailer" toml:"mailer"`
	Repos        []RepoOptions     `json:"repos" toml:"repos"`
	Packaging    *PackagingOptions `json:"packaging,omitempty" toml:"packaging,omitempty"`
}

type Host struct {
	Name string `json:"name,omitempty" toml:"name,omitempty"`
}

type MailerOptions struct {
	Type     string            `json:"type,omitempty" toml:"type,omitempty"` // smtp, gmail, webhook
	Host     string            `json:"host,omitempty" toml:"host,omitempty"`
	Port     int               `json:"port,omitempty" toml:"port,omitempty"`
	Username string            `json:"username,omitempty" toml:"username,omitempty"`
	Password string            `json:"password,omitempty" toml:"password,omitempty"`
	Meta     map[string]string `json:"meta,omitempty" toml:"meta,omitempty"`
}

type RepoOptions struct {
	URL  string `json:"url,omitempty" toml:"url,omitempty"`
	Type string `json:"type,omitempty" toml:"type,omitempty"` // http, embeded
	Slug string `json:"slug,omitempty" toml:"slug,omitempty"`
	Name string `json:"name,omitempty" toml:"name,omitempty"`
}

type PackagingOptions struct {
	OutputZipFile string   `json:"output_zip_file,omitempty" toml:"output_zip_file,omitempty"`
	IncludeFiles  []string `json:"include_files,omitempty" toml:"include_files,omitempty"`
	ExcludeFiles  []string `json:"exclude_files,omitempty" toml:"exclude_files,omitempty"`
}
