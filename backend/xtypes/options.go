package xtypes

type AppOptions struct {
	Name         string        `json:"name"`
	Port         int           `json:"port"`
	Host         string        `json:"host"`
	MasterSecret string        `json:"master_secret"`
	Debug        bool          `json:"debug"`
	WorkingDir   string        `json:"working_dir"`
	Mailer       MailerOptions `json:"mailer"`
}

type MailerOptions struct {
	Type     string            `json:"type"` // smtp, gmail, webhook
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Username string            `json:"username"`
	Password string            `json:"password"`
	Meta     map[string]string `json:"meta"`
}
