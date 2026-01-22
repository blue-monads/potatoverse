package executors

import (
	"os"

	"github.com/blue-monads/potatoverse/backend/xtypes"
)

type ExecState struct {
	SpaceId          int64
	PackageVersionId int64
	InstalledId      int64
	FsRoot           *os.Root
	App              xtypes.App
}
