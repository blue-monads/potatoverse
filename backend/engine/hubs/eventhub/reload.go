package eventhub

import "github.com/blue-monads/turnix/backend/utils/qq"

func (e *EventHub) watchReload() {

	for {
		select {
		case <-e.refreshFullIndex:
			qq.Println("@watchReload/refreshFullIndex")
			err := e.buildActiveEventsIndex()
			if err != nil {
				qq.Println("@watchReload/2")
			}
		}
	}

}
