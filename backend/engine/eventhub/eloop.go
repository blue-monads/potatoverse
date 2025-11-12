package eventhub

import (
	"time"

	"github.com/blue-monads/turnix/backend/utils/qq"
)

// fixme => add system for processing delayed targets

func (e *EventHub) eventLoop() {
	go e.rootEventWatcher()
	go e.eventProcessLoop()

	for range 10 {
		go e.targetProcessLoop()
	}

}

func (e *EventHub) rootEventWatcher() {

	fallbackTimer := time.NewTimer(time.Second * 30)

	checkForEvents := func() {
		events, err := e.sink.QueryNewEvents()
		if err != nil {
			qq.Println("@rootEventWatcher/checkForEvents/error", err)
		} else {
			for _, event := range events {
				e.eventProcessChan <- event
			}
		}
	}

	checkForTargets := func() {
		targets, err := e.sink.QueryNewEventTargets()
		if err != nil {
			qq.Println("@rootEventWatcher/QueryNewEventTargets/error", err)
		} else {
			for _, target := range targets {
				e.eventTargetProcessChan <- target
			}
		}
	}

	select {
	case <-fallbackTimer.C:
		checkForEvents()

		checkForTargets()

	case <-e.refreshFullIndex:
		qq.Println("@rootEventWatcher/refreshFullIndex")
		e.buildActiveEventsIndex()
	}

}

// new, scheduled, processed

func (e *EventHub) eventProcessLoop() {
	for eventId := range e.eventProcessChan {
		if eventId == 0 {
			continue
		}

		evt, err := e.sink.GetEvent(eventId)
		if err != nil {
			qq.Println("@eventProcessLoop/GetEvent/error", err)
			continue
		}

		if evt.Status == "processed" || evt.Status == "scheduled" {
			qq.Println("@eventProcessLoop/GetEvent/status", evt.Status)
			continue
		}

		targets, err := e.sink.CreateEventTargets(eventId)
		if err != nil {
			qq.Println("@eventProcessLoop/CreateEventTargets/error", err)
			continue
		}

		err = e.sink.UpdateEvent(eventId, map[string]any{
			"status": "scheduled",
		})

		if err != nil {
			qq.Println("@eventProcessLoop/UpdateEvent/error", err)
			continue
		}

		for _, targetId := range targets {
			e.eventTargetProcessChan <- targetId
		}

	}
}

func (e *EventHub) targetProcessLoop() {

	for targetId := range e.eventTargetProcessChan {
		if targetId == 0 {
			continue
		}

		err := e.targetProcessor(targetId)
		if err != nil {
			qq.Println("@targetProcessLoop/targetProcessor/error", err)
			continue
		}
	}

}
