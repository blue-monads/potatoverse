package eventhub

import (
	"context"
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

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewEventHub(db datahub.Database) *EventHub {
	ctx, cancel := context.WithCancel(context.Background())

	return &EventHub{
		sink:                   db.GetMQSynk(),
		db:                     db,
		activeEvents:           make(map[string]bool),
		activeEventsLock:       sync.RWMutex{},
		eventProcessChan:       make(chan int64, 13),
		eventTargetProcessChan: make(chan int64, 27),
		refreshFullIndex:       make(chan struct{}, 1),
		ctx:                    ctx,
		cancel:                 cancel,
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

	qq.Println("@Publish/1")

	if !e.needsProcessing(installId, name) {
		qq.Println("@Publish/2")
		return nil
	}

	qq.Println("@Publish/3")

	eventId, err := e.sink.AddEvent(installId, name, payload)
	if err != nil {
		qq.Println("@Publish/4")
		return err
	}

	qq.Println("@Publish/4")

	select {
	case e.eventProcessChan <- eventId:
	case <-e.ctx.Done():
		return e.ctx.Err()
	}

	qq.Println("@Publish/5")

	return nil
}

func (e *EventHub) RefreshFullIndex() {

	select {
	case e.refreshFullIndex <- struct{}{}:
	default:
	}
}

func (e *EventHub) buildActiveEventsIndex() error {
	sops := e.db.GetSpaceOps()

	subs, err := sops.QueryAllEventSubscriptions(false)
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

func (e *EventHub) Stop() {
	// Cancel context to signal all goroutines to stop
	e.cancel()

	// Wait for all goroutines to finish
	e.wg.Wait()
}
