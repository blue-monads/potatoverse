package executors

import (
	"os"

	"github.com/blue-monads/potatoverse/backend/engine/executors/luaz"
	"github.com/blue-monads/potatoverse/backend/registry"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

func init() {
	registry.RegisterExecutorBuilderFactory("luaz", luaz.BuildLuazExecutorBuilder)
}

type ExecState struct {
	SpaceId          int64
	PackageVersionId int64
	InstalledId      int64
	FsRoot           *os.Root
	App              xtypes.App
}
