package buddyhub

import (
	"log/slog"
	"net/http"
	"path"

	"github.com/blue-monads/potatoverse/backend/services/buddyhub/hq"
	"github.com/blue-monads/potatoverse/backend/utils/nostrutils"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
)

var (
	DefaultServers = []string{
		// "wss://relay.nostrich.de",
		// "wss://relay.snort.social",
		// "wss://nos.lol",
		"ws://local-hq.tubersalltheway.top",
	}
	_ IBuddyHub = (*BuddyHub)(nil)
)

type BuddyHub struct {
	hq     hq.HQ
	logger *slog.Logger
	app    xtypes.App

	baseBuddyDir string

	pubkey        string
	privkey       string
	port          int
	staticBuddies map[string]*xtypes.BuddyInfo
}

type Options struct {
	Logger             *slog.Logger
	App                xtypes.App
	DisableNostrServer bool
}

func NewBuddyHub(opt Options) *BuddyHub {

	config := opt.App.Config().(*xtypes.AppOptions)
	port := config.Port

	pubkey, pk, err := nostrutils.GenerateKeyPair(config.MasterSecret)
	if err != nil {
		opt.Logger.Error("Failed to generate key pair", "err", err)
		panic(err)
	}

	staticBuddies := make(map[string]*xtypes.BuddyInfo)

	if config.BuddyOptions != nil {
		for _, buddyInfo := range config.BuddyOptions.StaticBuddies {
			staticBuddies[buddyInfo.Pubkey] = buddyInfo
		}
	}

	buddyDir := path.Join(config.WorkingDir, "buddy")

	hqi, err := hq.New(hq.Options{
		Servers:    DefaultServers,
		PrivateKey: pk,
		PublicKey:  pubkey,
		Logger:     opt.Logger,
	})
	if err != nil {
		panic(err)
	}

	b := &BuddyHub{
		logger:        opt.Logger,
		app:           opt.App,
		baseBuddyDir:  buddyDir,
		hq:            *hqi,
		pubkey:        pubkey,
		privkey:       pk,
		port:          port,
		staticBuddies: staticBuddies,
	}

	b.pubkey = pubkey
	b.privkey = pk
	b.port = port

	return b
}

func (b *BuddyHub) GetPubkey() string {
	return b.pubkey
}

func (b *BuddyHub) GetPrivkey() string {
	return b.privkey
}

func (b *BuddyHub) Start() error {
	return b.hq.Start()
}

func (b *BuddyHub) ListBuddies() ([]*xtypes.BuddyInfo, error) {
	return nil, nil
}

func (b *BuddyHub) PingBuddy(buddyPubkey string) (bool, error) {
	return true, nil
}

func (b *BuddyHub) SendBuddy(buddyPubkey string, req *http.Request) (*http.Response, error) {
	return nil, nil
}

func (b *BuddyHub) RouteToBuddy(buddyPubkey string, ctx *gin.Context) {

}

func (b *BuddyHub) GetHQUrls() []string {
	return DefaultServers
}
