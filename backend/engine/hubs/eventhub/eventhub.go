package eventhub

import (
	"fmt"
	"sync"

	"github.com/blue-monads/turnix/backend/engine/hubs/eventhub/eslayer"
	"github.com/blue-monads/turnix/backend/services/datahub"
	"github.com/blue-monads/turnix/backend/utils/qq"
	"github.com/blue-monads/turnix/backend/xtypes"
)

type EventHub struct {
	app  xtypes.App
	sink datahub.MQSynk
	db   datahub.Database

	activeEvents     map[string]bool
	activeEventsLock sync.RWMutex

	refreshFullIndex chan struct{}

	eslayer *eslayer.ESLayer
}

func NewEventHub(app xtypes.App) *EventHub {
	db := app.Database()
	sink := db.GetMQSynk()

	return &EventHub{
		app:              app,
		sink:             sink,
		db:               db,
		activeEvents:     make(map[string]bool),
		activeEventsLock: sync.RWMutex{},
	}
}

func (e *EventHub) Start() error {
	err := e.buildActiveEventsIndex()
	if err != nil {
		qq.Println("@Start/buildActiveEventsIndex/error", err)
		return err
	}

	e.eslayer = eslayer.NewESLayer(e.app)
	err = e.eslayer.Start()
	if err != nil {
		qq.Println("@Start/eslayer.Start/error", err)
		return err
	}

	go e.watchReload()

	return nil
}

func (e *EventHub) Publish(opts *xtypes.EventOptions) error {
	installId := opts.InstallId
	name := opts.Name
	payload := opts.Payload

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

	e.eslayer.NotifyNewEvent(eventId)

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
	e.eslayer.Stop()
}
