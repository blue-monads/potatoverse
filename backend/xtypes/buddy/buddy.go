package buddy

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type BuddyInfo struct {
	Pubkey          string   `json:"pubkey"`
	URLs            []string `json:"urls"`
	AllowStorage    bool     `json:"allow_storage"`
	MaxStorage      int64    `json:"max_storage"`
	AllowWebFunnel  bool     `json:"allow_web_funnel"`
	MaxTrafficLimit int64    `json:"max_traffic_limit"`
}

type BuddyHub interface {
	GetPubkey() string
	GetPrivkey() string
	ListBuddies() ([]*BuddyInfo, error)
	PingBuddy(buddyPubkey string) (bool, error)
	SendBuddy(buddyPubkey string, req *http.Request) (*http.Response, error)
	RouteToBuddy(buddyPubkey string, ctx *gin.Context)
	GetBuddyRoot(buddyPubkey string) (*os.Root, error)

	// 	OpenBuddyWs(buddyPubkey string, endpoint string, inchan []byte, outchan chan []byte) error
}
