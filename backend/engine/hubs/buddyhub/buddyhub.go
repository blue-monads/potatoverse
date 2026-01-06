package buddyhub

import (
	"log/slog"
	"path"
	"sync"

	xutils "github.com/blue-monads/turnix/backend/utils"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
	"github.com/blue-monads/turnix/backend/xtypes/buddy"
)

type BuddyHub struct {
	logger         *slog.Logger
	app            xtypes.App
	buddyDir       string
	rendezvousUrls []string

	pendingRequests     map[string]chan *buddy.Response
	pendingRequestsLock sync.RWMutex

	allowAllBuddies bool

	allbuddyAllowStorage bool
	allbuddyMaxStorage   int64

	allbuddyAllowWebFunnel  bool
	allbuddyMaxTrafficLimit int64

	staticBuddies map[string]*buddy.BuddyInfo

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
		logger:                  opt.Logger,
		app:                     opt.App,
		buddyDir:                buddyDir,
		pendingRequests:         make(map[string]chan *buddy.Response),
		pendingRequestsLock:     sync.RWMutex{},
		allowAllBuddies:         false,
		allbuddyAllowStorage:    false,
		allbuddyMaxStorage:      0,
		allbuddyAllowWebFunnel:  false,
		allbuddyMaxTrafficLimit: 0,
		staticBuddies:           make(map[string]*buddy.BuddyInfo),
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

func (h *BuddyHub) Ping(providerURL string) (bool, error) {
	return true, nil
}

func (h *BuddyHub) PingBuddy(providerURL string, buddyPubkey string) (bool, error) {
	return true, nil
}
