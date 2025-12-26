package eslayer

import (
	"time"

	"github.com/blue-monads/turnix/backend/utils/qq"
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
	e.wg.Add(1)
	defer e.wg.Done()

	fallbackTimer := time.NewTimer(time.Second * 30)
	defer fallbackTimer.Stop()

	checkForEvents := func() {
		events, err := e.sink.QueryNewEvents()
		if err != nil {
			qq.Println("@rootEventWatcher/checkForEvents/error", err)
		} else {
			for _, event := range events {
				select {
				case e.eventProcessChan <- event:
				case <-e.ctx.Done():
					return
				}
			}
		}
	}

	checkForTargets := func() {
		targets, err := e.sink.QueryNewEventTargets()
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
		targets, err := e.sink.QueryDelayExpiredTargets()
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

	for {
		// Reset timer for next iteration
		fallbackTimer.Reset(time.Second * 30)

		select {
		case <-e.ctx.Done():
			return
		case <-fallbackTimer.C:
			checkForEvents()
			checkForTargets()
			checkForDelayedTargets()

		}

	}

}

// new, scheduled, processed

func (e *ESLayer) eventProcessLoop() {
	e.wg.Add(1)
	defer e.wg.Done()

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
				select {
				case e.eventTargetProcessChan <- targetId:
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
