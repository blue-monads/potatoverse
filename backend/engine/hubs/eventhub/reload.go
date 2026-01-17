package eventhub

import (
	"github.com/blue-monads/potatoverse/backend/utils/qq"
)

func (e *EventHub) watchReload() {

	for event := range e.refreshFullIndex {

		qq.Println("@watchReload/refreshFullIndex", event)

		err := e.buildActiveEventsIndex()
		if err != nil {
			qq.Println("@watchReload/2", err)
		}

	}

}
