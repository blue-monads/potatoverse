package xtypes

type AppOptions struct {
	Name         string           `json:"name,omitempty" yaml:"name,omitempty"`
	Port         int              `json:"port,omitempty" yaml:"port,omitempty"`
	Hosts        []Host           `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	MasterSecret string           `json:"master_secret,omitempty" yaml:"master_secret,omitempty"`
	Debug        bool             `json:"debug_mode,omitempty" yaml:"debug_mode,omitempty"`
	WorkingDir   string           `json:"working_dir,omitempty" yaml:"working_dir,omitempty"`
	SocketFile   string           `json:"socket_file,omitempty" yaml:"socket_file,omitempty"`
	Mailer       MailerOptions    `json:"mailer" yaml:"mailer"`
	Repos        []RepoOptions    `json:"repos" yaml:"repos"`
	BuddyOptions *BuddyHubOptions `json:"buddy_options,omitempty" yaml:"buddy_options,omitempty"`
}

type Host struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

type MailerOptions struct {
	Type     string            `json:"type,omitempty" yaml:"type,omitempty"` // smtp, gmail, webhook
	Host     string            `json:"host,omitempty" yaml:"host,omitempty"`
	Port     int               `json:"port,omitempty" yaml:"port,omitempty"`
	Username string            `json:"username,omitempty" yaml:"username,omitempty"`
	Password string            `json:"password,omitempty" yaml:"password,omitempty"`
	Meta     map[string]string `json:"meta,omitempty" yaml:"meta,omitempty"`
}

type RepoOptions struct {
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
	Type string `json:"type,omitempty" yaml:"type,omitempty"` // http, embeded
	Slug string `json:"slug,omitempty" yaml:"slug,omitempty"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

type BuddyHubOptions struct {
	AllowAllBuddies         bool            `json:"allow_all_buddies,omitempty" yaml:"allow_all_buddies,omitempty"`
	AllBuddyAllowStorage    bool            `json:"all_buddy_allow_storage,omitempty" yaml:"all_buddy_allow_storage,omitempty"`
	AllBuddyMaxStorage      int64           `json:"all_buddy_max_storage,omitempty" yaml:"all_buddy_max_storage,omitempty"`
	AllBuddyMaxTrafficLimit int64           `json:"all_buddy_max_traffic_limit,omitempty" yaml:"all_buddy_max_traffic_limit,omitempty"`
	BuddyWebFunnelMode      string          `json:"buddy_web_funnel_mode,omitempty" yaml:"buddy_web_funnel_mode,omitempty"`
	StaticBuddies           []*BuddyInfo    `json:"static_buddies,omitempty" yaml:"static_buddies,omitempty"`
	RendezvousUrls          []RendezvousUrl `json:"rendezvous_urls,omitempty" yaml:"rendezvous_urls,omitempty"`
}

type BuddyInfo struct {
	Name            string     `json:"name"`
	Pubkey          string     `json:"pubkey"`
	URLs            []BuddyUrl `json:"urls"`
	AllowStorage    bool       `json:"allow_storage"`
	MaxStorage      int64      `json:"max_storage"`
	AllowWebFunnel  bool       `json:"allow_web_funnel"`
	MaxTrafficLimit int64      `json:"max_traffic_limit"`
}

type BuddyUrl struct {
	Host       string `json:"host"`
	Port       string `json:"port"`
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
