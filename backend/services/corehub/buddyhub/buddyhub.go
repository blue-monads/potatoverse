package buddyhub

import (
	"log/slog"
	"net/http"
	"os"
	"path"

	"github.com/blue-monads/turnix/backend/services/corehub/buddyhub/funnel"
	xutils "github.com/blue-monads/turnix/backend/utils"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/buddy"
	"github.com/gin-gonic/gin"
)

var (
	_ buddy.BuddyHub = (*BuddyHub)(nil)
)

type Configuration struct {
	allowAllBuddies bool

	allbuddyAllowStorage bool
	allbuddyMaxStorage   int64

	buddyAllowWebFunnelMode string // funnel_mode (all, local, none)
	allbuddyMaxTrafficLimit int64
}

type BuddyHub struct {
	logger         *slog.Logger
	app            xtypes.App
	baseBuddyDir   string
	rendezvousUrls []string

	configuration Configuration

	staticBuddies map[string]*buddy.BuddyInfo

	funnel *funnel.Funnel

	pubkey  string
	privkey string
}

type Options struct {
	Logger             *slog.Logger
	App                xtypes.App
	DisableNostrServer bool
}

func NewBuddyHub(opt Options) *BuddyHub {

	config := opt.App.Config().(*xtypes.AppOptions)

	buddyDir := path.Join(config.WorkingDir, "buddy")
	b := &BuddyHub{
		logger:       opt.Logger,
		app:          opt.App,
		baseBuddyDir: buddyDir,

		configuration: Configuration{
			allowAllBuddies:         false,
			allbuddyAllowStorage:    false,
			allbuddyMaxStorage:      0,
			buddyAllowWebFunnelMode: "none",
			allbuddyMaxTrafficLimit: 0,
		},

		staticBuddies: make(map[string]*buddy.BuddyInfo),
	}

	pubkey, pk, err := xutils.GenerateKeyPair(config.MasterSecret)
	if err != nil {
		b.logger.Error("Failed to generate key pair", "err", err)
		panic(err)
	}

	b.pubkey = pubkey
	b.privkey = pk

	qq.Println("@pubkey", pubkey)

	return b
}

func (h *BuddyHub) GetPubkey() string {
	return h.pubkey
}

func (h *BuddyHub) GetPrivkey() string {
	return h.privkey
}

func (h *BuddyHub) ListBuddies() ([]*buddy.BuddyInfo, error) {
	return nil, nil
}

func (h *BuddyHub) PingBuddy(buddyPubkey string) (bool, error) {
	return true, nil
}

func (h *BuddyHub) SendBuddy(buddyPubkey string, req *http.Request) (*http.Response, error) {
	return nil, nil
}

func (h *BuddyHub) RouteToBuddy(buddyPubkey string, ctx *gin.Context) {
	h.funnel.HandleRoute(buddyPubkey, ctx)
}

func (h *BuddyHub) GetBuddyRoot(buddyPubkey string) (*os.Root, error) {
	return nil, nil
}

func (h *BuddyHub) GetRendezvousUrls() []buddy.RendezvousUrl {
	return nil
}
