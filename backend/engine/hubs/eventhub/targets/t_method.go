package targets

import (
	"fmt"

	"github.com/blue-monads/potatoverse/backend/engine"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/evtype"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/blue-monads/potatoverse/backend/xtypes/lazydata"
	"github.com/blue-monads/potatoverse/backend/xtypes/xcapability/easyaction"
)

func PerformSpaceMethodTargetExecution(app xtypes.App) evtype.Handler {

	engine := app.Engine().(*engine.Engine)

	return func(execution *evtype.TExecution) error {
		targetSpaceId := execution.Subscription.TargetSpaceID
		if targetSpaceId == 0 {
			execution.RetryAble = false
			return fmt.Errorf("target space id is required")
		}

		err := engine.EmitActionEvent(&xtypes.ActionEventOptions{
			SpaceId:    execution.Subscription.SpaceID,
			EventType:  "event_target",
			ActionName: execution.Subscription.TargetEndpoint,
			Params: map[string]string{
				"event_id":        fmt.Sprintf("%d", execution.Event.ID),
				"target_id":       fmt.Sprintf("%d", execution.Subscription.ID),
				"target_type":     execution.Subscription.TargetType,
				"origin_space_id": fmt.Sprintf("%d", execution.Subscription.SpaceID),
			},
			Request: &MQActionContext{
				exec: execution,
			},
		})
		if err != nil {
			execution.RetryAble = false
			return fmt.Errorf("error emitting action event: %w", err)
		}

		return nil

	}

}

type MQActionContext struct {
	exec *evtype.TExecution
}

func (c *MQActionContext) ListActions() ([]string, error) {
	return easyaction.Methods, nil
}

func (c *MQActionContext) ExecuteAction(name string, params lazydata.LazyData) (any, error) {
	return easyaction.BytelazyDataActions(c.exec.Event.Payload, name, params)
}
