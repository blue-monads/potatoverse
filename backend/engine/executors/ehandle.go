package executors

import (
	"log/slog"
	"os"

	"github.com/blue-monads/turnix/backend/xtypes"
)

type EHandle struct {
	Logger  *slog.Logger
	App     xtypes.App
	FsRoot  *os.Root
	SpaceId int64
}
