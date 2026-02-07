package eslayer

import (
	"fmt"
	"time"

	"github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/evtype"
	"github.com/blue-monads/potatoverse/backend/engine/hubs/eventhub/rengine"
	"github.com/blue-monads/potatoverse/backend/utils/qq"
)

func (e *ESLayer) targetProcessor(targetId int64) error {

	qq.Println("targetProcessor/1")

	sink := e.datahandle.GetMQSynk()
	sops := e.datahandle.GetSpaceOps()

	qq.Println("targetProcessor/2")

	target, err := sink.TransitionTargetStart(targetId)
	if err != nil {
		qq.Println("targetProcessor/3", err)
		return err
	}

	qq.Println("targetProcessor/4")

	event, err := sink.GetEvent(target.EventID)
	if err != nil {
		qq.Println("targetProcessor/5", err)
		return err
	}

	qq.Println("targetProcessor/6")

	sub, err := sops.GetEventSubscription(event.InstallID, target.SubscriptionID)
	if err != nil {
		qq.Println("targetProcessor/7", err)
		return err
	}

	qq.Println("targetProcessor/8", sub)

	ok, err := rengine.RuleEngine(sub.Rules, event.Payload)
	if err != nil {
		qq.Println("targetProcessor/9", err)
		sink.TransitionTargetFail(event.ID, targetId, err.Error())
		return err
	}
	if !ok {
		qq.Println("targetProcessor/10", "rules not matched")
		return nil
	}

	handler, ok := e.handlers[sub.TargetType]
	if !ok {
		qq.Println("targetProcessor/11", "handler not found", sub.TargetType)
		return fmt.Errorf("handler not found: %s", sub.TargetType)
	}

	ectx := &evtype.TExecution{
		Subscription: sub,
		Target:       target,
		Event:        event,
	}

	if sub.DelayStart > 0 && target.Status != "start_delayed" {
		delayStart := time.Now().Unix() + sub.DelayStart*1000
		err = sink.TransitionTargetStartDelayed(targetId, event.ID, delayStart)
		if err != nil {
			qq.Println("targetProcessor/11", err)
			sink.TransitionTargetFail(event.ID, targetId, err.Error())
			return err
		}

		qq.Println("targetProcessor/12", "delayed for", sub.DelayStart, "seconds")
		return nil
	}

	err = handler(ectx)
	if err != nil {
		qq.Println("targetProcessor/12", err)

		sink.TransitionTargetFail(event.ID, targetId, err.Error())
		return err
	}
	return nil
}
