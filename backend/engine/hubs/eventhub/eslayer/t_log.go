package eslayer

import (
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/tidwall/pretty"
)

func PerformLogTargetExecution(ectx *TargetExecution) error {
	qq.Println("PerformLogTargetExecution", ectx.Event.Payload)
	result := pretty.Color(ectx.Event.Payload, nil)
	qq.Println("PerformLogTargetExecution/1", string(result))

	return nil
}
