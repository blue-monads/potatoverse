package targets

import "github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/evtype"

func init() {
	evtype.RegisterTargetBuilder("log", PerformLogTargetExecution)
	evtype.RegisterTargetBuilder("method", PerformSpaceMethodTargetExecution)
	evtype.RegisterTargetBuilder("script", PerformScriptTargetExecution)
	evtype.RegisterTargetBuilder("webhook", PerformWebhookTargetExecution)
}
