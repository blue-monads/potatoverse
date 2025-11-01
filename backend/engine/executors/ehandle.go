package executors

import (
	"log/slog"
	"os"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/xtypes"
)

type EHandle struct {
	Logger           *slog.Logger
	App              xtypes.App
	FsRoot           *os.Root
	SpaceId          int64
	PackageVersionId int64
	InstalledId      int64
	Database         datahub.Database
}
