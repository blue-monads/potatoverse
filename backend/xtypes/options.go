package xtypes

type AppOptions struct {
	Name         string        `json:"name,omitempty" toml:"name,omitempty"`
	Port         int           `json:"port,omitempty" toml:"port,omitempty"`
	Host         string        `json:"host,omitempty" toml:"host,omitempty"`
	MasterSecret string        `json:"master_secret,omitempty" toml:"master_secret,omitempty"`
	Debug        bool          `json:"debug_mode,omitempty" toml:"debug_mode,omitempty"`
	WorkingDir   string        `json:"working_dir,omitempty" toml:"working_dir,omitempty"`
	SocketFile   string        `json:"socket_file,omitempty" toml:"socket_file,omitempty"`
	Mailer       MailerOptions `json:"mailer,omitempty" toml:"mailer,omitempty"`
}

type MailerOptions struct {
	Type     string            `json:"type,omitempty" toml:"type,omitempty"` // smtp, gmail, webhook
	Host     string            `json:"host,omitempty" toml:"host,omitempty"`
	Port     int               `json:"port,omitempty" toml:"port,omitempty"`
	Username string            `json:"username,omitempty" toml:"username,omitempty"`
	Password string            `json:"password,omitempty" toml:"password,omitempty"`
	Meta     map[string]string `json:"meta,omitempty" toml:"meta,omitempty"`
}
