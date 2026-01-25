package buddyhub

import (
	"log/slog"

	"github.com/blue-monads/potatoverse/backend/services/buddyhub/hq"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

type BuddyHub struct {
	hq     hq.HQ
	logger *slog.Logger
	app    xtypes.App
}
