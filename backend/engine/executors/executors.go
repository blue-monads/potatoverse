package executors

import (
	"github.com/blue-monads/turnix/backend/engine/executors/luaz"
	"github.com/blue-monads/turnix/backend/engine/registry"
)

func init() {
	registry.RegisterExecutorBuilderFactory("luaz", luaz.BuildLuazExecutorBuilder)
}
