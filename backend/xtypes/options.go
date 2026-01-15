package xtypes

type AppOptions struct {
	Name         string           `json:"name,omitempty" toml:"name,omitempty"`
	Port         int              `json:"port,omitempty" toml:"port,omitempty"`
	Hosts        []Host           `json:"hosts,omitempty" toml:"hosts,omitempty"`
	MasterSecret string           `json:"master_secret,omitempty" toml:"master_secret,omitempty"`
	Debug        bool             `json:"debug_mode,omitempty" toml:"debug_mode,omitempty"`
	WorkingDir   string           `json:"working_dir,omitempty" toml:"working_dir,omitempty"`
	SocketFile   string           `json:"socket_file,omitempty" toml:"socket_file,omitempty"`
	Mailer       MailerOptions    `json:"mailer" toml:"mailer"`
	Repos        []RepoOptions    `json:"repos" toml:"repos"`
	BuddyOptions *BuddyHubOptions `json:"buddy_options,omitempty" toml:"buddy_options,omitempty"`
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

type BuddyHubOptions struct {
	AllowAllBuddies         bool            `json:"allow_all_buddies,omitempty" toml:"allow_all_buddies,omitempty"`
	AllBuddyAllowStorage    bool            `json:"all_buddy_allow_storage,omitempty" toml:"all_buddy_allow_storage,omitempty"`
	AllBuddyMaxStorage      int64           `json:"all_buddy_max_storage,omitempty" toml:"all_buddy_max_storage,omitempty"`
	AllBuddyMaxTrafficLimit int64           `json:"all_buddy_max_traffic_limit,omitempty" toml:"all_buddy_max_traffic_limit,omitempty"`
	BuddyWebFunnelMode      string          `json:"buddy_web_funnel_mode,omitempty" toml:"buddy_web_funnel_mode,omitempty"`
	StaticBuddies           []*BuddyInfo    `json:"static_buddies,omitempty" toml:"static_buddies,omitempty"`
	RendezvousUrls          []RendezvousUrl `json:"rendezvous_urls,omitempty" toml:"rendezvous_urls,omitempty"`
}

type BuddyInfo struct {
	Pubkey          string     `json:"pubkey"`
	URLs            []BuddyUrl `json:"urls"`
	AllowStorage    bool       `json:"allow_storage"`
	MaxStorage      int64      `json:"max_storage"`
	AllowWebFunnel  bool       `json:"allow_web_funnel"`
	MaxTrafficLimit int64      `json:"max_traffic_limit"`
}

type BuddyUrl struct {
	Endpoint   string `json:"endpoint"`
	IsDefault  bool   `json:"is_default"`
	Priority   int    `json:"priority"`
	Provider   string `json:"provider"` // funnel, direct, nostr, udp, libp2p(lpweb), tor etc
	PreConnect bool   `json:"pre_connect"`
}

type RendezvousUrl struct {
	URL        string `json:"url"`
	Provider   string `json:"provider"` // nostr, udp, libp2p, tor etc
	Priority   int    `json:"priority"`
	SimpleMode bool   `json:"simple_mode"`
}
