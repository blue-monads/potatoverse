package eslayer

import (
	"github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/evtype"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/tidwall/pretty"
)

func PerformLogTargetExecution(app xtypes.App) func(ectx *evtype.TExecution) error {

	return func(ectx *evtype.TExecution) error {
		qq.Println("PerformLogTargetExecution", ectx.Event.Payload)
		result := pretty.Color(ectx.Event.Payload, nil)
		qq.Println("PerformLogTargetExecution/1", string(result))
		return nil
	}

}
