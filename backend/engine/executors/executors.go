package executors

import (
	"github.com/blue-monads/potatoverse/backend/engine/executors/luaz"
	"github.com/blue-monads/potatoverse/backend/registry"
)

func init() {
	registry.RegisterExecutorBuilderFactory("luaz", luaz.BuildLuazExecutorBuilder)
}
