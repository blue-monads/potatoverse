package buddytypes

import (
	"net/http"
	"os"

	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type BuddyHub interface {
	GetPubkey() string
	GetPrivkey() string
	ListBuddies() ([]*xtypes.BuddyInfo, error)
	PingBuddy(buddyPubkey string) (bool, error)
	SendBuddy(buddyPubkey string, req *http.Request) (*http.Response, error)
	RouteToBuddy(buddyPubkey string, ctx *gin.Context)
	GetBuddyRoot(buddyPubkey string) (*os.Root, error)
	GetRendezvousUrls() []xtypes.RendezvousUrl
}
