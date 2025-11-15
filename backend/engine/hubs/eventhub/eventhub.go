package eventhub

import (
	"fmt"
	"sync"

	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/utils/qq"
)

type EventHub struct {
	sink datahub.MQSynk
	db   datahub.Database

	activeEvents     map[string]bool
	activeEventsLock sync.RWMutex

	refreshFullIndex chan struct{}

	eventProcessChan       chan int64
	eventTargetProcessChan chan int64
}

func NewEventHub(db datahub.Database) *EventHub {

	return &EventHub{
		sink:                   nil,
		activeEvents:           make(map[string]bool),
		activeEventsLock:       sync.RWMutex{},
		eventProcessChan:       make(chan int64, 13),
		eventTargetProcessChan: make(chan int64, 27),
		refreshFullIndex:       make(chan struct{}, 1),
	}
}

func (e *EventHub) Start() error {
	err := e.buildActiveEventsIndex()
	if err != nil {
		qq.Println("@Start/buildActiveEventsIndex/error", err)
		return err
	}

	e.eventLoop()

	return nil
}

func (e *EventHub) Publish(installId int64, name string, payload []byte) error {

	if !e.needsProcessing(installId, name) {
		return nil
	}

	eventId, err := e.sink.AddEvent(installId, name, payload)
	if err != nil {
		return err
	}

	e.eventProcessChan <- eventId

	qq.Println("@published/event", eventId)

	return nil
}

func (e *EventHub) RefreshFullIndex() {
	e.refreshFullIndex <- struct{}{}
}

func (e *EventHub) buildActiveEventsIndex() error {
	sops := e.db.GetSpaceOps()

	subs, err := sops.QueryAllEventSubscriptions()
	if err != nil {
		return err
	}

	nextIndex := make(map[string]bool)
	for _, sub := range subs {
		nextIndex[fmt.Sprintf("%d||%s", sub.InstallID, sub.EventKey)] = true
	}

	e.activeEventsLock.Lock()
	defer e.activeEventsLock.Unlock()

	e.activeEvents = nextIndex

	return nil

}

func (e *EventHub) needsProcessing(installId int64, name string) bool {
	key := fmt.Sprintf("%d||%s", installId, name)

	e.activeEventsLock.RLock()
	defer e.activeEventsLock.RUnlock()

	return e.activeEvents[key]
}
