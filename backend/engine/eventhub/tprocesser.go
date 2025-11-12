package eventhub

import (
	"time"

	"github.com/blue-monads/turnix/backend/utils/qq"
)

func (e *EventHub) targetProcessor(targetId int64) error {
	qq.Println("@targetProcessor/start")

	time.Sleep(time.Second * 10)

	qq.Println("@targetProcessor/end")

	return nil

}
