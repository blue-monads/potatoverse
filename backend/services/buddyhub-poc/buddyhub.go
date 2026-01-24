package buddyhub

import (
	"log/slog"
	"net/http"
	"os"
	"path"

	buddy "github.com/blue-monads/potatoverse/backend/services/buddyhub-poc/buddytypes"
	"github.com/blue-monads/potatoverse/backend/services/buddyhub-poc/funnel"
	xutils "github.com/blue-monads/potatoverse/backend/utils"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
)

var (
	_ buddy.BuddyHub = (*BuddyHub)(nil)
)

type Configuration struct {
	allowAllBuddies bool

	allbuddyAllowStorage bool
	allbuddyMaxStorage   int64

	buddyWebFunnelMode      string // funnel_mode (all, local, none)
	allbuddyMaxTrafficLimit int64

	rendezvousUrls []xtypes.RendezvousUrl
}

type BuddyHub struct {
	logger       *slog.Logger
	app          xtypes.App
	baseBuddyDir string

	configuration Configuration

	staticBuddies map[string]*xtypes.BuddyInfo

	funnel *funnel.Funnel

	pubkey  string
	privkey string
	port    int
}

type Options struct {
	Logger             *slog.Logger
	App                xtypes.App
	DisableNostrServer bool
}

func NewBuddyHub(opt Options) *BuddyHub {

	config := opt.App.Config().(*xtypes.AppOptions)

	port := config.Port

	buddyDir := path.Join(config.WorkingDir, "buddy")
	b := &BuddyHub{
		logger:       opt.Logger,
		app:          opt.App,
		baseBuddyDir: buddyDir,

		configuration: Configuration{
			allowAllBuddies:         false,
			allbuddyAllowStorage:    false,
			allbuddyMaxStorage:      0,
			buddyWebFunnelMode:      "none",
			allbuddyMaxTrafficLimit: 0,
		},

		staticBuddies: make(map[string]*xtypes.BuddyInfo),
	}

	pubkey, pk, err := xutils.GenerateKeyPair(config.MasterSecret)
	if err != nil {
		b.logger.Error("Failed to generate key pair", "err", err)
		panic(err)
	}

	b.pubkey = pubkey
	b.privkey = pk
	b.port = port

	b.funnel = funnel.New()

	err = b.configure(config)
	if err != nil {
		b.logger.Error("Failed to configure buddy hub", "err", err)
		panic(err)
	}

	b.startRloop()

	qq.Println("@pubkey", pubkey)

	return b
}

func (h *BuddyHub) GetPubkey() string {
	return h.pubkey
}

func (h *BuddyHub) GetPrivkey() string {
	return h.privkey
}

func (h *BuddyHub) ListBuddies() ([]*xtypes.BuddyInfo, error) {

	result := make([]*xtypes.BuddyInfo, 0, len(h.staticBuddies))

	for _, buddyInfo := range h.staticBuddies {
		result = append(result, buddyInfo)
	}

	return result, nil
}

func (h *BuddyHub) PingBuddy(buddyPubkey string) (bool, error) {
	return true, nil
}

func (h *BuddyHub) SendBuddy(buddyPubkey string, req *http.Request) (*http.Response, error) {
	return nil, nil
}

func (h *BuddyHub) HandleFunnelRoute(buddyPubkey string, ctx *gin.Context) {

	buddyInfo, exists := h.staticBuddies[buddyPubkey]
	if !exists {
		return
	}

	if !buddyInfo.AllowWebFunnel {
		return
	}

	h.funnel.HandleServerWebSocket(buddyPubkey, ctx)
}

func (h *BuddyHub) RouteToBuddy(buddyPubkey string, ctx *gin.Context) {
	h.funnel.HandleRoute(buddyPubkey, ctx)
}

func (h *BuddyHub) GetBuddyRoot(buddyPubkey string) (*os.Root, error) {
	return nil, nil
}

func (h *BuddyHub) GetRendezvousUrls() []xtypes.RendezvousUrl {
	return nil
}

func (h *BuddyHub) RegisterHandler(msgType string, handler func(buddyPubkey string, data []byte) error) error {
	return nil
}
