package eventhub

import (
	"time"

	"github.com/blue-monads/turnix/backend/utils/qq"
)

func (e *EventHub) targetProcessor(targetId int64) error {

	qq.Println("targetProcessor/1")

	sops := e.db.GetSpaceOps()

	qq.Println("targetProcessor/2")

	target, err := e.sink.StartTargetProcess(targetId)
	if err != nil {
		qq.Println("targetProcessor/3", err)
		return err
	}

	qq.Println("targetProcessor/4")

	event, err := e.sink.GetEvent(target.EventID)
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

	time.Sleep(time.Second * 10)

	err = e.sink.CompleteTargetProcess(targetId)
	if err != nil {
		qq.Println("targetProcessor/9", err)
		return err
	}

	return nil

}
