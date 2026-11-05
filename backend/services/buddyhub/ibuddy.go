package buddyhub

import (
	"net/http"

	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type IBuddyHub interface {
	GetHQTunnelDomain() string
	GetPrivkey() string
	GetPubkey() string

	HandleFunnelRegisterNode(buddyPubkey string, ctx *gin.Context)
	HandleFunnelRoute(buddyPubkey string, ctx *gin.Context)

	ListBuddies() ([]*xtypes.BuddyInfo, error)

	SendBuddy(buddyPubkey string, req *http.Request) (*http.Response, error)

	Start() error
	Stop() error
}

func NewDummyBuddyHub() IBuddyHub {
	return &DummyBuddyHub{}
}

var _ IBuddyHub = (*DummyBuddyHub)(nil)

type DummyBuddyHub struct{}

func (dbh *DummyBuddyHub) GetHQTunnelDomain() string {
	return "dummy-buddyhub.local"
}

func (dbh *DummyBuddyHub) GetPrivkey() string {
	return "dummy-privkey"
}

func (dbh *DummyBuddyHub) GetPubkey() string {
	return "dummy-pubkey"
}

func (dbh *DummyBuddyHub) HandleFunnelRegisterNode(buddyPubkey string, ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "funnel register node handled"})
}

func (dbh *DummyBuddyHub) HandleFunnelRoute(buddyPubkey string, ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "funnel route handled"})
}

func (dbh *DummyBuddyHub) ListBuddies() ([]*xtypes.BuddyInfo, error) {
	return []*xtypes.BuddyInfo{}, nil
}

func (dbh *DummyBuddyHub) SendBuddy(buddyPubkey string, req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
	}, nil
}

func (dbh *DummyBuddyHub) Start() error {
	return nil
}

func (dbh *DummyBuddyHub) Stop() error {
	return nil
}
