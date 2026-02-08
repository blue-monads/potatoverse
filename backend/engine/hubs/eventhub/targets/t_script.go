package eslayer

import (
	"github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/evtype"
	"github.com/blue-monads/potatoverse/backend/xtypes"
)

func PerformScriptTargetExecution(app xtypes.App) func(execution *evtype.TExecution) error {

	return func(execution *evtype.TExecution) error {
		return nil
	}

	return nil
}
