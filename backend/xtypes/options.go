package xtypes

type AppOptions struct {
	Name         string        `json:"name" toml:"name"`
	Port         int           `json:"port" toml:"port"`
	Host         string        `json:"host" toml:"host"`
	MasterSecret string        `json:"master_secret" toml:"master_secret"`
	Debug        bool          `json:"debug_mode" toml:"debug_mode"`
	WorkingDir   string        `json:"working_dir" toml:"working_dir"`
	SocketFile   string        `json:"socket_file" toml:"socket_file"`
	Mailer       MailerOptions `json:"mailer" toml:"mailer"`
}

type MailerOptions struct {
	Type     string            `json:"type" toml:"type"` // smtp, gmail, webhook
	Host     string            `json:"host" toml:"host"`
	Port     int               `json:"port" toml:"port"`
	Username string            `json:"username" toml:"username"`
	Password string            `json:"password" toml:"password"`
	Meta     map[string]string `json:"meta" toml:"meta"`
}
