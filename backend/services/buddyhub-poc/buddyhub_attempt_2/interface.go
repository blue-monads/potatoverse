package buddyhub

import (
	"net/http"

	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type IBuddyHub interface {
	Start() error
	GetPubkey() string
	GetPrivkey() string
	ListBuddies() ([]*xtypes.BuddyInfo, error)
	PingBuddy(buddyPubkey string) (bool, error)
	SendBuddy(buddyPubkey string, req *http.Request) (*http.Response, error)
	RouteToBuddy(buddyPubkey string, ctx *gin.Context)
	GetHQUrls() []string
}
