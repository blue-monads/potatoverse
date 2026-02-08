package eslayer

import (
	"fmt"
	"time"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/evtype"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/rengine"
	qq "github.com/blue-monads/potatoverse/backend/utils/qq"
)

func (e *ESLayer) targetProcessor(targetId int64) error {
	qq.Println("@targetProcessor: processing target", targetId)

	sink := e.datahandle.GetMQSynk()
	sops := e.datahandle.GetSpaceOps()

	target, err := sink.TransitionTargetStart(targetId)
	if err != nil {
		qq.Println("@targetProcessor/TransitionTargetStart/error", err)
		return err
	}

	event, err := sink.GetEvent(target.EventID)
	if err != nil {
		qq.Println("@targetProcessor/GetEvent/error", err)
		return err
	}

	sub, err := sops.GetEventSubscription(event.InstallID, target.SubscriptionID)
	if err != nil {
		qq.Println("@targetProcessor/GetEventSubscription/error", err)
		return err
	}

	ok, err := rengine.RuleEngine(sub.Rules, event.Payload)
	if err != nil {
		sink.TransitionTargetFail(event.ID, targetId, err.Error())
		qq.Println("@targetProcessor/RuleEngine/error", err)
		return err
	}
	if !ok {
		qq.Println("@targetProcessor/RuleEngine: no match")
		return nil
	}

	handler, ok := e.handlers[sub.TargetType]
	if !ok {
		err = fmt.Errorf("handler not found: %s", sub.TargetType)
		qq.Println("@targetProcessor/handler-not-found", err)
		return err
	}

	ectx := &evtype.TExecution{
		Subscription: sub,
		Target:       target,
		Event:        event,
	}

	if sub.DelayStart > 0 && target.Status != "start_delayed" {
		delayStart := time.Now().Unix() + int64(sub.DelayStart)
		err = sink.TransitionTargetStartDelayed(targetId, event.ID, delayStart)
		if err != nil {
			sink.TransitionTargetFail(event.ID, targetId, err.Error())
			qq.Println("@targetProcessor/TransitionTargetStartDelayed/error", err)
			return err
		}
		qq.Println("@targetProcessor: target", targetId, "delayed until", delayStart)
		return nil
	}

	qq.Println("@targetProcessor: calling handler for target", targetId)
	err = handler(ectx)
	if err != nil {
		qq.Println("@targetProcessor/handler/error", err)
		// Check if this is a retryable error
		if ectx.RetryAble && sub.MaxRetries > 0 && target.RetryCount < sub.MaxRetries {
			newRetryCount := target.RetryCount + 1
			delayUntil := time.Now().Unix() + int64(sub.RetryDelay)

			err = sink.TransitionTargetDelay(targetId, event.ID, delayUntil, newRetryCount)
			if err != nil {
				sink.TransitionTargetFail(event.ID, targetId, err.Error())
				return err
			}
			qq.Println("@targetProcessor: target", targetId, "retried, new count", newRetryCount)
			return nil
		}

		sink.TransitionTargetFail(event.ID, targetId, err.Error())
		return err
	}

	// Mark target as completed
	err = sink.TransitionTargetComplete(event.ID, targetId)
	if err != nil {
		qq.Println("@targetProcessor/TransitionTargetComplete/error", err)
		return err
	}

	qq.Println("@targetProcessor: target", targetId, "completed")
	return nil
}
