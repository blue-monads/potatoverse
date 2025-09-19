package bhandle

import (
	"log/slog"
	"os"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/services/signer"
	"github.com/blue-monads/turnix/backend/xtypes"
)

type Bhandle struct {
	Database datahub.Database
	Logger   *slog.Logger
	Signer   *signer.Signer
	App      xtypes.App
	FsRoot   *os.Root
	SpaceId  int64
}
