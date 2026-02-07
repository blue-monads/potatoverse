package buddyhub

import (
	"fmt"
	"log/slog"
	"net/http"
	"path"

	"github.com/blue-monads/potatoverse/backend/services/buddyhub/funnel"
	"github.com/blue-monads/potatoverse/backend/utils/nostrutils"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type Options struct {
	Logger *slog.Logger
	App    xtypes.App
}

type BuddyHub struct {
	funnelHQ *funnel.FunnelClient

	logger *slog.Logger

	baseBuddyDir string

	pubkey        string
	privkey       string
	port          int
	staticBuddies map[string]*xtypes.BuddyInfo
}

const (
	DefaultFunnelHQ = "ws://localhost:7447"
)

func NewBuddyHub(config *xtypes.AppOptions, logger *slog.Logger) *BuddyHub {

	port := config.Port

	pubkey, pk, err := nostrutils.GenerateKeyPair(config.MasterSecret)
	if err != nil {
		logger.Error("Failed to generate key pair", "err", err)
		panic(err)
	}

	bh := &BuddyHub{
		logger:        logger,
		funnelHQ:      nil,
		baseBuddyDir:  path.Join(config.WorkingDir, "buddy"),
		pubkey:        pubkey,
		privkey:       pk,
		port:          port,
		staticBuddies: make(map[string]*xtypes.BuddyInfo),
	}

	if config.BuddyOptions != nil {
		for _, buddyInfo := range config.BuddyOptions.StaticBuddies {
			bh.staticBuddies[buddyInfo.Pubkey] = buddyInfo
		}
	}

	bh.funnelHQ = funnel.NewFunnelClient(funnel.FunnelClientOptions{
		LocalHttpPort:   port,
		RemoteFunnelUrl: DefaultFunnelHQ,
		NodeId:          pubkey,
	})

	return bh
}

func (bh *BuddyHub) Start() error {
	return bh.funnelHQ.Start(bh.pubkey)
}

func (bh *BuddyHub) Stop() error {
	return nil
}

func (bh *BuddyHub) GetPubkey() string {
	return bh.pubkey
}

func (bh *BuddyHub) GetPrivkey() string {
	return bh.privkey
}

func (bh *BuddyHub) ListBuddies() ([]*xtypes.BuddyInfo, error) {
	return nil, nil
}

func (bh *BuddyHub) PingBuddy(buddyPubkey string) (bool, error) {
	return true, nil
}

func (bh *BuddyHub) SendBuddy(buddyPubkey string, req *http.Request) (*http.Response, error) {

	buddyInfo, exists := bh.staticBuddies[buddyPubkey]
	if !exists {
		return nil, fmt.Errorf("buddy not found: %s", buddyPubkey)
	}

	for _, url := range buddyInfo.URLs {
		provider := url.Provider
		if provider != "http" {
			continue
		}

		req.URL.Host = fmt.Sprintf("%s:%s", url.Host, url.Port)
		req.URL.Scheme = "http"

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		return resp, nil

	}

	return nil, fmt.Errorf("no provider found for buddy: %s", buddyPubkey)
}

func (bh *BuddyHub) HandleFunnelRoute(buddyPubkey string, ctx *gin.Context) {

	// fixme run emebed funnel server

}
