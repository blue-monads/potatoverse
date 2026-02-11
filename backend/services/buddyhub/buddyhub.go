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

	embeddedFunnel *funnel.Funnel
}

const (
	DefaultFunnelHQ      = "https://tubersalltheway.top/zz/buddy/register"
	EnableEmbeddedFunnel = true
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

	token, err := nostrutils.GenerateNostrAuthToken(bh.privkey, DefaultFunnelHQ, "GET")
	if err != nil {
		logger.Error("Failed to generate nostr auth token", "err", err)
		panic(err)
	}

	finalUrl := fmt.Sprintf("%s?token=%s", DefaultFunnelHQ, token)

	bh.funnelHQ = funnel.NewFunnelClient(funnel.FunnelClientOptions{
		LocalHttpPort:   port,
		RemoteFunnelUrl: finalUrl,
		NodeId:          pubkey,
	})

	return bh
}

func (bh *BuddyHub) Start() error {
	err := bh.funnelHQ.Start(bh.pubkey)
	if err != nil {
		return err
	}

	if EnableEmbeddedFunnel {
		bh.embeddedFunnel = funnel.New()
	}

	return nil
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
	result := make([]*xtypes.BuddyInfo, 0, len(bh.staticBuddies))
	for _, buddyInfo := range bh.staticBuddies {
		result = append(result, buddyInfo)
	}

	return result, nil
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
	if bh.embeddedFunnel == nil {
		return
	}

	bh.embeddedFunnel.HandleServerWebSocket(buddyPubkey, ctx)

}

func (bh *BuddyHub) HandleFunnelRegisterNode(buddyPubkey string, ctx *gin.Context) {
	if bh.embeddedFunnel == nil {
		return
	}

	bh.embeddedFunnel.HandleServerWebSocket(buddyPubkey, ctx)
}
