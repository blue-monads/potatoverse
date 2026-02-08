package eslayer

import (
	"time"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
)

// fixme => add system for processing delayed targets

func (e *ESLayer) eventLoop() {
	go e.rootEventWatcher()
	go e.eventProcessLoop()

	for range 10 {
		go e.targetProcessLoop()
	}

}

func (e *ESLayer) rootEventWatcher() {

	sink := e.datahandle.GetMQSynk()

	checkForEvents := func() {
		events, err := sink.QueryNewEvents()
		if err != nil {
			qq.Println("@rootEventWatcher/checkForEvents/error", err)
		} else {
			qq.Println("@rootEventWatcher/checkForEvents: found", len(events), "new events")
			for _, event := range events {
				select {
				case e.eventProcessChan <- event:
					qq.Println("@rootEventWatcher/checkForEvents: sent event", event)
				case <-e.ctx.Done():
					return
				}
			}
		}
	}

	checkForTargets := func() {
		targets, err := sink.QueryNewEventTargets()
		if err != nil {
			qq.Println("@rootEventWatcher/QueryNewEventTargets/error", err)
		} else {
			for _, target := range targets {
				select {
				case e.eventTargetProcessChan <- target:
				case <-e.ctx.Done():
					return
				}
			}
		}
	}

	checkForDelayedTargets := func() {
		targets, err := sink.QueryDelayExpiredTargets()
		if err != nil {
			qq.Println("@rootEventWatcher/QueryDelayExpiredTargets/error", err)
		} else {
			for _, target := range targets {
				select {
				case e.eventTargetProcessChan <- target:
				case <-e.ctx.Done():
					return
				}
			}
		}
	}

	fallbackTimer := time.NewTimer(time.Second * 2)
	defer fallbackTimer.Stop()

	counter := 0

	for {

		qq.Println("@couner/start", counter)

		// Reset timer for next iteration
		fallbackTimer.Reset(time.Second * 2)

		select {
		case <-e.ctx.Done():
			return
		case <-fallbackTimer.C:

			qq.Println("@checkForEvents")
			checkForEvents()

			qq.Println("@checkForTargets")
			checkForTargets()

			qq.Println("@checkForDelayedTargets")
			checkForDelayedTargets()

		}

		qq.Println("@couner/end", counter)

		counter = counter + 1

	}

}

// new, scheduled, processed

func (e *ESLayer) eventProcessLoop() {
	e.wg.Add(1)
	defer e.wg.Done()

	sink := e.datahandle.GetMQSynk()

	for {
		select {
		case <-e.ctx.Done():
			return
		case eventId, ok := <-e.eventProcessChan:
			if !ok {
				return
			}
			if eventId == 0 {
				continue
			}
			qq.Println("@eventProcessLoop: received event", eventId)

			evt, err := sink.GetEvent(eventId)
			if err != nil {
				qq.Println("@eventProcessLoop/GetEvent/error", err)
				continue
			}

			if evt.Status == "processed" || evt.Status == "scheduled" {
				qq.Println("@eventProcessLoop/GetEvent/status", evt.Status)
				continue
			}

			qq.Println("@eventProcessLoop: creating targets for event", eventId)
			targets, err := sink.CreateEventTargets(eventId)
			if err != nil {
				qq.Println("@eventProcessLoop/CreateEventTargets/error", err)
				continue
			}

			err = sink.UpdateEvent(eventId, map[string]any{
				"status": "scheduled",
			})

			if err != nil {
				qq.Println("@eventProcessLoop/UpdateEvent/error", err)
				continue
			}

			qq.Println("@eventProcessLoop: created", len(targets), "targets for event", eventId)
			for _, targetId := range targets {
				select {
				case e.eventTargetProcessChan <- targetId:
					qq.Println("@eventProcessLoop: sent target", targetId)
				case <-e.ctx.Done():
					return
				}
			}
		}
	}
}

func (e *ESLayer) targetProcessLoop() {
	e.wg.Add(1)
	defer e.wg.Done()

	for {
		select {
		case <-e.ctx.Done():
			return
		case targetId, ok := <-e.eventTargetProcessChan:
			if !ok {
				return
			}
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

}
