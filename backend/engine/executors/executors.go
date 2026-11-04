package executors

import (
	"os"

	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type ExecState struct {
	SpaceId          int64
	PackageVersionId int64
	InstalledId      int64
	FsRoot           *os.Root
	App              xtypes.App
	Closable         *Closable
	Executor         xtypes.Executor
	HttpRequest      *gin.Context
}

func (es *ExecState) Init() {
	es.Closable = &Closable{}
}
