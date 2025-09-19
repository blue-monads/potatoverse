package bhandle

import (
	"log/slog"
	"os"

	"github.com/blue-monads/turnix/backend/xtypes"
)

type Bhandle struct {
	Logger  *slog.Logger
	App     xtypes.App
	FsRoot  *os.Root
	SpaceId int64
}
