package eventhub

import (
	"github.com/blue-monads/potatoverse/backend/utils/qq"
)

func (e *EventHub) watchReload() {

	for range e.refreshFullIndex {

		qq.Println("@watchReload/refreshFullIndex")

		err := e.buildActiveEventsIndex()
		if err != nil {
			qq.Println("@watchReload/2", err)
		}

	}

}
