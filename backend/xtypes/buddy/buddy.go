package buddy

import "encoding/json"

type BuddyInfo struct {
	URL             string   `json:"url"`
	AltURLs         []string `json:"alt_urls"`
	Pubkey          string   `json:"pubkey"`
	AllowStorage    bool     `json:"allow_storage"`
	MaxStorage      int64    `json:"max_storage"`
	AllowWebFunnel  bool     `json:"allow_web_funnel"`
	MaxTrafficLimit int64    `json:"max_traffic_limit"`
}

type Message struct {
	MType     string          `json:"mtype"`
	Payload   json.RawMessage `json:"payload"`
	RequestId string          `json:"request_id"`
}

type Response struct {
	RequestId string          `json:"request_id"`
	Status    string          `json:"status"`
	Data      json.RawMessage `json:"data"`
}

type BuddySyncProvider interface {
	Ping(providerURL string) (bool, error)
	PingBuddy(providerURL string, buddyPubkey string) (bool, error)
	SendBuddy(providerURL string, buddyPubkey string, message *Message) (*Response, error)
}
