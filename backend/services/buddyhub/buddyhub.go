package buddyhub

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/blue-monads/potatoverse/backend/services/buddyhub/funnel"
	"github.com/blue-monads/potatoverse/backend/utils/nostrutils"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type Options struct {
	Logger *slog.Logger
	App    xtypes.App
}

type BuddyHub struct {
	logger *slog.Logger

	baseBuddyDir string

	pubkey        string
	privkey       string
	port          int
	staticBuddies map[string]*xtypes.BuddyInfo

	embeddedFunnel *funnel.Funnel

	hqURl string
}

const (
	CloudFunnelURL = "http://tubersalltheway.top/zz/buddy/register"
	LocalFunnelURL = "http://localhost:7771/zz/buddy/register"

	DefaultFunnelHQ = CloudFunnelURL
)

func NewBuddyHub(config *xtypes.AppOptions, logger *slog.Logger) *BuddyHub {

	port := config.Port

	pubkey, pk, err := nostrutils.GenerateKeyPair(config.MasterSecret)
	if err != nil {
		logger.Error("Failed to generate key pair", "err", err)
		panic(err)
	}

	funnelHQ := DefaultFunnelHQ

	envHq := os.Getenv("POTATO_DEFAULT_HQ")

	if envHq != "" {
		funnelHQ = envHq
	}

	bh := &BuddyHub{
		logger:        logger,
		baseBuddyDir:  path.Join(config.WorkingDir, "buddy"),
		pubkey:        pubkey,
		privkey:       pk,
		port:          port,
		staticBuddies: make(map[string]*xtypes.BuddyInfo),
		hqURl:         funnelHQ,
	}

	if config.BuddyOptions != nil {
		for _, buddyInfo := range config.BuddyOptions.StaticBuddies {
			bh.staticBuddies[buddyInfo.Pubkey] = buddyInfo
		}
	}

	return bh
}

func (bh *BuddyHub) Start() error {

	token, err := nostrutils.GenerateNostrAuthToken(bh.privkey, DefaultFunnelHQ, "GET")
	if err != nil {
		bh.logger.Error("Failed to generate nostr auth token", "err", err)
		panic(err)
	}

	go func() {

		for {
			funnelHQ := funnel.NewFunnelClient(funnel.FunnelClientOptions{
				LocalHttpPort:   bh.port,
				RemoteFunnelUrl: bh.hqURl,
				NodeId:          bh.pubkey,
			})

			err = funnelHQ.Start(token)
			if err != nil {
				qq.Println("@err", err.Error())
			}

			funnelHQ.Stop()

			time.Sleep(time.Second * 20)

		}

	}()

	if os.Getenv("POTATO_DISABLE_EMBED_FUNNEL") != "1" {
		bh.embeddedFunnel = funnel.New()
		err := bh.embeddedFunnel.StartKcpServer()
		if err != nil {
			bh.logger.Error("Failed to start KCP server", "err", err)
		}
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

func (bh *BuddyHub) GetHQTunnelDomain() string {

	hqurl, err := url.Parse(bh.hqURl)
	if err != nil {
		return ""
	}

	nodeId := nostrutils.PubKeyToNodeId(bh.pubkey)
	final := fmt.Sprintf("buddy-%s.%s", nodeId, hqurl.Host)

	return final

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

	bh.embeddedFunnel.HandleRoute(buddyPubkey, ctx)

}

func (bh *BuddyHub) HandleFunnelRegisterNode(buddyPubkey string, ctx *gin.Context) {
	if bh.embeddedFunnel == nil {
		return
	}

	bh.embeddedFunnel.HandleServerWebSocket(buddyPubkey, ctx)
}
