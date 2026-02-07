package buddyhub

import (
	"log/slog"
	"net/http"
	"path"

	"github.com/blue-monads/potatoverse/backend/services/buddyhub/funnel"
	"github.com/blue-monads/potatoverse/backend/utils/nostrutils"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

type Options struct {
	Logger *slog.Logger
	App    xtypes.App
}

type BuddyHub struct {
	funnelHQ *funnel.FunnelClient

	logger *slog.Logger
	app    xtypes.App

	baseBuddyDir string

	pubkey        string
	privkey       string
	port          int
	staticBuddies map[string]*xtypes.BuddyInfo
}

const (
	DefaultFunnelHQ = "ws://localhost:7447"
)

func NewBuddyHub(opt Options) *BuddyHub {

	config := opt.App.Config().(*xtypes.AppOptions)
	port := config.Port

	pubkey, pk, err := nostrutils.GenerateKeyPair(config.MasterSecret)
	if err != nil {
		opt.Logger.Error("Failed to generate key pair", "err", err)
		panic(err)
	}

	bh := &BuddyHub{
		logger:        opt.Logger,
		app:           opt.App,
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
	return nil, nil
}
